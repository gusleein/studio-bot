package database

// Tx — интерфейс транзакции базы данных.
type Tx interface {
	Commit() error
	Rollback() error
}
