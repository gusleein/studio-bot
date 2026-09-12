package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// txKey — тип ключа для хранения транзакции в контексте.
type txKey struct{}

// WithTx оборачивает функцию fn в транзакцию.
// При ошибке fn транзакция откатывается; при панике — тоже откатывается и паника пробрасывается.
func WithTx(ctx context.Context, db *sqlx.DB, fn func(ctx context.Context) error) error {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("BeginTxx: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err = fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("fn error: %w; rollback error: %v", err, rbErr)
		}
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}

	return nil
}

// TxFromContext извлекает транзакцию из контекста.
// Возвращает nil, false, если транзакции нет.
func TxFromContext(ctx context.Context) (*sqlx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sqlx.Tx)
	return tx, ok
}

// GetDB возвращает активную транзакцию из контекста (если есть),
// иначе — основной пул соединений db.
func GetDB(ctx context.Context, db *sqlx.DB) DB {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}
	return db
}
