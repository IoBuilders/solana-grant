package db

type Transaction interface {
	Commit() error
	Rollback() error
}
