package transactions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

func ProcessListSqlite(db *sql.DB, accountID string, txs []Transaction) error {
	ctx := context.Background()
	conn, _ := db.Conn(ctx)
	defer func() {
		_ = conn.Close()
	}()

	_, err := conn.ExecContext(ctx, "BEGIN EXCLUSIVE TRANSACTION;")
	if err != nil {
		return fmt.Errorf("could not start transaction: %v", err)
	}

	insertSQL := `INSERT INTO transactions (account, id, amount_cents, description) VALUES ($1, $2, $3, $4)`

	var total int64
	for _, tx := range txs {
		transactionID := uuid.New().String()
		_, err := conn.ExecContext(ctx, insertSQL, accountID, transactionID, tx.AmountCents, tx.Description)
		if err != nil {
			if _, rErr := conn.ExecContext(ctx, "ROLLBACK"); rErr != nil {
				return rollbackSqlite(ctx, conn, fmt.Errorf("could not insert transaction: %v", err))
			}
		}

		total += tx.AmountCents
	}

	_, err = conn.ExecContext(
		ctx,
		`UPDATE accounts SET balance_cents = balance_cents - $1 WHERE id = $2`,
		total, accountID,
	)
	if err != nil {
		return rollbackSqlite(ctx, conn, fmt.Errorf("could not update balance: %v", err))
	}

	var balance int64
	row := conn.QueryRowContext(ctx, "SELECT balance_cents FROM accounts WHERE id = $1", accountID)
	if err := row.Scan(&balance); err != nil {
		return rollbackSqlite(ctx, conn, fmt.Errorf("could not scan balance: %v", err))
	}

	if balance < 0 {
		return rollbackSqlite(ctx, conn, fmt.Errorf("insufficient funds %.2f on account ID %d", float64(balance/100), accountID))
	}

	if _, err = conn.ExecContext(ctx, "COMMIT"); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	return nil
}

func rollbackSqlite(ctx context.Context, conn *sql.Conn, wrappingError error) error {
	if _, err := conn.ExecContext(ctx, "ROLLBACK"); err != nil {
		return fmt.Errorf("could not rollback (wrapped error: %v): %v", wrappingError, err)
	} else {
		return wrappingError
	}
}
