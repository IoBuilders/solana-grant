package utils

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"dlt-ingress/src/main/config"
)

// testDBURL is populated by TestMain with a real PostgreSQL DSN.
// Tests that need a database call requireDB(t) which skips when this is empty.
var testDBURL string

// TestMain starts a testcontainers PostgreSQL instance shared by the whole
// package test binary. If TEST_DATABASE_URL is set, it is used as-is.
// If Docker is unavailable, the database-dependent tests are skipped gracefully
// rather than failing the whole binary (so unit tests such as
// TestEnsureTimezoneUTC still pass).
func TestMain(m *testing.M) {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		testDBURL = url
		os.Exit(m.Run())
		return
	}

	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx,
		"postgres:14.7",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		fmt.Printf("WARNING: skipping DB integration tests (could not start postgres: %v)\n", err)
		os.Exit(m.Run())
		return
	}
	defer func() { _ = pgContainer.Terminate(ctx) }()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Printf("WARNING: skipping DB integration tests (connection string unavailable: %v)\n", err)
		os.Exit(m.Run())
		return
	}
	testDBURL = connStr
	os.Exit(m.Run())
}

// requireDB skips the calling test when no database is available.
func requireDB(t *testing.T) {
	t.Helper()
	if testDBURL == "" {
		t.Skip("no database available: set TEST_DATABASE_URL or ensure Docker is running")
	}
}

// TestNewPostgresDB_ConnectTimeout_IsEnforced starts a TCP listener that
// accepts the connection but never sends any PostgreSQL protocol data.
// The driver must give up after connect_timeout seconds.
// No real PostgreSQL instance is required for this test.
//
// Why a silent TCP listener: it separates the TCP-reachability signal
// (connect succeeds immediately) from the PostgreSQL-handshake signal
// (never arrives), so the driver's connect_timeout — not the OS TCP timeout —
// is what terminates the attempt.
func TestNewPostgresDB_ConnectTimeout_IsEnforced(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	// connReceived carries the accepted connection so t.Cleanup can close it.
	// Buffered so the goroutine never blocks on the send.
	connReceived := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			connReceived <- conn
		}
		close(connReceived)
	}()

	// Close listener first (unblocks Accept if no connection arrived yet),
	// then close any connection that was accepted.
	t.Cleanup(func() {
		err := listener.Close()
		if err != nil {
			return
		}
		if conn, ok := <-connReceived; ok {
			err := conn.Close()
			if err != nil {
				return
			}
		}
	})

	cfg := &config.DatabaseConfig{
		Url:            fmt.Sprintf("postgres://testuser:testpass@%s/testdb?sslmode=disable", listener.Addr()),
		ConnectTimeout: 2 * time.Second,
	}

	start := time.Now()

	// gorm.Open with database/sql is lazy; Ping forces the first real connection.
	db, openErr := NewPostgresDB(cfg)
	if openErr == nil {
		sqlDB, sqlErr := db.DB()
		if sqlErr == nil {
			// Use a generous context so we are measuring the driver-level
			// connect_timeout (2 s), not a context deadline.
			pingCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			openErr = sqlDB.PingContext(pingCtx)
			_ = sqlDB.Close()
		} else {
			openErr = sqlErr
		}
	}

	elapsed := time.Since(start)

	require.Error(t, openErr, "expected connect timeout error, got nil")
	// The exact error message is driver- and OS-dependent; timing is the
	// authoritative signal that connect_timeout was actually applied.
	assert.GreaterOrEqual(t, elapsed, 1*time.Second,
		"connection failed before connect_timeout; timeout may not be configured (elapsed: %v)", elapsed)
	assert.Less(t, elapsed, 5*time.Second,
		"connect_timeout did not fire within the expected window (elapsed: %v)", elapsed)
}

// TestNewPostgresDB_StatementTimeout_IsEnforced runs a pg_sleep(4) query
// against a DB configured with a 2-second statement_timeout and asserts that
// PostgreSQL cancels the query with the expected error before it completes.
func TestNewPostgresDB_StatementTimeout_IsEnforced(t *testing.T) {
	requireDB(t)

	cfg := &config.DatabaseConfig{
		Url:                testDBURL,
		StatementTimeout:   2 * time.Second,
		MaxOpenConnections: 5,
		MaxIdleConnections: 2,
	}

	db, err := NewPostgresDB(cfg)
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	start := time.Now()
	result := db.Exec("SELECT pg_sleep(4)")
	elapsed := time.Since(start)

	require.Error(t, result.Error, "expected statement timeout error, got nil")
	assert.Contains(t, result.Error.Error(), "canceling statement due to statement timeout",
		"unexpected error: %v", result.Error)
	assert.GreaterOrEqual(t, elapsed, 1*time.Second,
		"cancelled before timeout; statement_timeout may not be configured (elapsed: %v)", elapsed)
	assert.Less(t, elapsed, 3500*time.Millisecond,
		"not cancelled within expected window (elapsed: %v)", elapsed)
}

// TestNewPostgresDB_LockTimeout_IsEnforced verifies that a transaction waiting
// to acquire a row lock held by another transaction is cancelled after the
// configured lock_timeout duration.
//
// Concurrency protocol:
//  1. Goroutine A acquires a FOR UPDATE lock on a dedicated test row and
//     signals success via lockResult (buffered chan error).
//  2. The main goroutine waits for that signal, then issues the contending
//     UPDATE — it will block until the lock_timeout fires.
//  3. txBDone is closed (sync.Once) once the UPDATE returns, releasing
//     goroutine A so it can roll back and exit.
//  4. t.Cleanup closes txBDone as a safety net if the test exits early via
//     a failed require, preventing goroutine leaks.
func TestNewPostgresDB_LockTimeout_IsEnforced(t *testing.T) {
	requireDB(t)

	cfg := &config.DatabaseConfig{
		Url:                testDBURL,
		LockTimeout:        2 * time.Second,
		MaxOpenConnections: 10,
		MaxIdleConnections: 5,
	}

	db, err := NewPostgresDB(cfg)
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	// Unique table name prevents interference between concurrent test runs.
	tableName := "test_lock_" + sanitizeIdentifier(t.Name())
	ctx := context.Background()

	_, err = sqlDB.ExecContext(ctx, fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (id INT PRIMARY KEY, val TEXT)", tableName,
	))
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = sqlDB.ExecContext(context.Background(), "DROP TABLE IF EXISTS "+tableName)
	})

	_, err = sqlDB.ExecContext(ctx, fmt.Sprintf(
		"INSERT INTO %s (id, val) VALUES (1, 'initial') ON CONFLICT (id) DO NOTHING", tableName,
	))
	require.NoError(t, err)

	// txBDone is closed exactly once when the UPDATE in the main goroutine
	// completes (or when t.Cleanup runs as a safety net).
	txBDone := make(chan struct{})
	var txBDoneOnce sync.Once
	signalTxBDone := func() { txBDoneOnce.Do(func() { close(txBDone) }) }
	t.Cleanup(signalTxBDone)

	lockResult := make(chan error, 1)

	// Goroutine A: acquire the row lock on a dedicated sql.Conn so the pool
	// cannot reclaim the connection while the lock is held.
	go func() {
		connA, err := sqlDB.Conn(ctx)
		if err != nil {
			lockResult <- fmt.Errorf("conn: %w", err)
			return
		}
		defer func(connA *sql.Conn) {
			err := connA.Close()
			if err != nil {
				lockResult <- fmt.Errorf("conn close: %w", err)
				return
			}
		}(connA)

		txA, err := connA.BeginTx(ctx, nil)
		if err != nil {
			lockResult <- fmt.Errorf("begin: %w", err)
			return
		}
		defer txA.Rollback() //nolint:errcheck

		_, err = txA.ExecContext(ctx, fmt.Sprintf(
			"SELECT id FROM %s WHERE id = 1 FOR UPDATE", tableName,
		))
		if err != nil {
			lockResult <- fmt.Errorf("select for update: %w", err)
			return
		}

		lockResult <- nil // row is locked; notify main goroutine
		<-txBDone         // hold the lock until transaction B finishes
	}()

	require.NoError(t, <-lockResult, "goroutine A failed to acquire the row lock")

	// Transaction B: UPDATE the locked row — must block until lock_timeout fires.
	// Because lock_timeout is embedded in the DSN, every pooled connection
	// carries the setting from the moment it is established.
	start := time.Now()
	result := db.Exec(fmt.Sprintf(
		"UPDATE %s SET val = 'updated' WHERE id = 1", tableName,
	))
	elapsed := time.Since(start)

	signalTxBDone()

	require.Error(t, result.Error, "expected lock timeout error, got nil")
	assert.Contains(t, result.Error.Error(), "canceling statement due to lock timeout",
		"unexpected error: %v", result.Error)
	assert.GreaterOrEqual(t, elapsed, 1500*time.Millisecond,
		"lock_timeout fired too quickly (elapsed: %v)", elapsed)
	assert.Less(t, elapsed, 5*time.Second,
		"lock_timeout did not fire within the expected window (elapsed: %v)", elapsed)
}

// TestNewPostgresDB_IdleInTransactionSessionTimeout_IsEnforced opens a
// transaction, executes one query (confirming it is healthy), then lets it
// sit idle for longer than idle_in_transaction_session_timeout. The next
// query on the same transaction must fail because PostgreSQL has terminated
// the backend.
func TestNewPostgresDB_IdleInTransactionSessionTimeout_IsEnforced(t *testing.T) {
	requireDB(t)

	cfg := &config.DatabaseConfig{
		Url:                             testDBURL,
		IdleInTransactionSessionTimeout: 2 * time.Second,
		MaxOpenConnections:              5,
		MaxIdleConnections:              2,
	}

	db, err := NewPostgresDB(cfg)
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	tx, err := sqlDB.Begin()
	require.NoError(t, err)
	defer tx.Rollback() //nolint:errcheck

	// First query proves the transaction is open and the connection is healthy.
	_, err = tx.Exec("SELECT 1")
	require.NoError(t, err, "initial query should succeed before idle timeout")

	// The transaction is now idle-in-transaction. PostgreSQL starts the
	// idle_in_transaction_session_timeout clock after the query above completes.
	// Sleeping for 3 s guarantees the 2 s timeout has fired.
	time.Sleep(3 * time.Second)

	// PostgreSQL has terminated the backend; the next statement must fail.
	_, err = tx.Exec("SELECT 1")

	require.Error(t, err, "expected idle-in-transaction timeout error, got nil")
	// The exact error string varies by driver: pgx surfaces the PG error message;
	// lib/pq may return EOF or "bad connection" if it reads the closed socket first.
	errStr := strings.ToLower(err.Error())
	isExpected := strings.Contains(errStr, "idle-in-transaction") ||
		strings.Contains(errStr, "terminating connection") ||
		strings.Contains(errStr, "eof") ||
		strings.Contains(errStr, "bad connection")
	assert.True(t, isExpected,
		"error should indicate the backend was terminated, got: %v", err)
}

// TestNewPostgresDB_ConnectionPool_IsConfigured checks that MaxOpenConnections
// is reflected in sql.DBStats and that MaxIdleConnections is enforced: after
// releasing more connections than the idle limit the pool must not keep the
// excess ones alive.
func TestNewPostgresDB_ConnectionPool_IsConfigured(t *testing.T) {
	requireDB(t)

	const (
		maxOpen = 7
		maxIdle = 3
	)

	cfg := &config.DatabaseConfig{
		Url:                testDBURL,
		MaxOpenConnections: maxOpen,
		MaxIdleConnections: maxIdle,
	}

	db, err := NewPostgresDB(cfg)
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	assert.Equal(t, maxOpen, sqlDB.Stats().MaxOpenConnections,
		"MaxOpenConnections not reflected in sql.DBStats")

	// Open maxOpen dedicated connections so the pool is fully saturated,
	// then release them all. Connections beyond maxIdle are closed by the
	// pool rather than kept in the idle set.
	ctx := context.Background()
	conns := make([]interface{ Close() error }, maxOpen)
	for i := range conns {
		conn, err := sqlDB.Conn(ctx)
		require.NoError(t, err, "failed to acquire connection %d", i)
		conns[i] = conn
	}
	for _, conn := range conns {
		require.NoError(t, conn.Close())
	}

	assert.LessOrEqual(t, sqlDB.Stats().Idle, maxIdle,
		"idle connection count exceeds MaxIdleConnections after releasing %d connections", maxOpen)
}

// sanitizeIdentifier maps a test name to a lower-case, alphanumeric-only
// string suitable for use as a PostgreSQL table identifier.
func sanitizeIdentifier(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + 32)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}
