package health

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestChecker struct {
	CheckerName  string
	Status       Status
	ResponseTime time.Duration
	Error        string
}

func NewTestChecker(name string, status Status, responseTime time.Duration, err string) *TestChecker {
	return &TestChecker{
		CheckerName:  name,
		Status:       status,
		ResponseTime: responseTime,
		Error:        err,
	}
}

func (c *TestChecker) Name() string {
	return c.CheckerName
}

func (c *TestChecker) Check(ctx context.Context) CheckResult {
	return CheckResult{
		Status:       c.Status,
		ResponseTime: c.ResponseTime.String(),
		Error:        c.Error,
	}
}

func TestCoreHealthRegistry_Register(t *testing.T) {
	reg := NewRegistry()
	checker := NewTestChecker("test", StatusUp, 100*time.Millisecond, "")
	reg.Register(checker)

	assert.Equal(t, len(reg.checkers), 1)
	assert.Equal(t, reg.checkers[0].Name(), checker.Name())
}

func TestCoreHealthRegistry_RegisterConcurrently(t *testing.T) {
	reg := NewRegistry()
	var wg sync.WaitGroup

	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			checker := NewTestChecker("test ", StatusUp, 100*time.Millisecond, "")
			reg.Register(checker)
		}()
	}
	wg.Wait()

	assert.Equal(t, len(reg.checkers), 25)
}

func TestCoreHealthRegistry_RunAllChecks_Up(t *testing.T) {
	reg := NewRegistry()
	checker := NewTestChecker("test", StatusUp, 100*time.Millisecond, "")
	checker2 := NewTestChecker("test2", StatusUp, 200*time.Millisecond, "")
	reg.Register(checker)
	reg.Register(checker2)

	res := reg.RunAllChecks(context.Background())
	assert.Equal(t, res.Status, StatusUp)
	assert.NotNil(t, res.Timestamp)
	assert.Equal(t, len(res.Checks), 2)
	assert.Equal(t, res.Checks["test"].Status, checker.Status)
	assert.Equal(t, res.Checks["test"].ResponseTime, checker.ResponseTime.String())
	assert.Equal(t, res.Checks["test"].Error, checker.Error)
	assert.Equal(t, res.Checks["test2"].Status, checker2.Status)
	assert.Equal(t, res.Checks["test2"].ResponseTime, checker2.ResponseTime.String())
	assert.Equal(t, res.Checks["test2"].Error, checker2.Error)
}

func TestCoreHealthRegistry_RunAllChecks_Down(t *testing.T) {
	reg := NewRegistry()
	checker := NewTestChecker("test", StatusUp, 100*time.Millisecond, "")
	checker2 := NewTestChecker("test2", StatusDown, 200*time.Millisecond, "test error")
	reg.Register(checker)
	reg.Register(checker2)

	res := reg.RunAllChecks(context.Background())
	assert.Equal(t, res.Status, StatusDown)
	assert.NotNil(t, res.Timestamp)
	assert.Equal(t, len(res.Checks), 2)
	assert.Equal(t, res.Checks["test"].Status, checker.Status)
	assert.Equal(t, res.Checks["test"].ResponseTime, checker.ResponseTime.String())
	assert.Equal(t, res.Checks["test"].Error, checker.Error)
	assert.Equal(t, res.Checks["test2"].Status, checker2.Status)
	assert.Equal(t, res.Checks["test2"].ResponseTime, checker2.ResponseTime.String())
	assert.Equal(t, res.Checks["test2"].Error, checker2.Error)
}
