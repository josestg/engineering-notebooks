package sqltx

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// Transaction knows how to do query transactions.
// The Transaction doesn't have access to any data or query to be executed in
// the transaction statement, but we can provide them by using
// the High Order Function.
//
// Example:
//	func InsertUser(u *model.User) Transaction {
//		// here we can inject data by using Transaction creator.
//		// returns the Transaction with the implementation.
//		return func(ctx context.Context, tx *sqlx.Tx) (context.Context, error) {
//			if d == nil {
//				return ctx, nil
//			}
//
//			const q = `
//			INSERT INTO users(
//				id, name, email, created_at, updated_at, deleted_at)
//			VALUES ($1, $2, $3, $4, $5, $6);
//		`
//
//			_, err := tx.ExecContext(ctx, q,
//				u.ID,
//				u.name,
//				d.email,
//				d.CreatedAt,
//				d.UpdatedAt,
//				d.DeletedAt,
//			)
//			if err != nil {
//				return ctx, err
//			}
//
//			return ctx, nil
//		}
//	}
//
// The InsertUser also called as Transaction creator.
type Transaction func(ctx context.Context, tx *sqlx.Tx) (context.Context, error)

// ExecTransactions executes the given transactions in order.
// The reason why each execution returns a context is to provide some metadata
// to the next transaction calls. It can be useful for tracing.
//
// If any transaction is failed, all transactions will be rolling back.
func ExecTransactions(ctx context.Context, db *sqlx.DB, transactions ...Transaction) error {
	tx, err := db.Beginx()
	if err != nil {
		return errors.Wrap(err, "begin transaction")
	}

	for i := 0; i < len(transactions); i++ {
		if transactions[i] == nil {
			continue
		}

		// don't used `:=`, because we need to replace ctx with the returned
		// ctx to next calls.
		ctx, err = transactions[i](ctx, tx)

		// if one of transaction cause error, it should be rollback.
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return errors.Wrapf(err, "rolling back transactions[%d] failed with error: %v", i, rollbackErr)
			}

			return errors.Wrapf(err, "evaluating transactions[%d]", i)
		}
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}
