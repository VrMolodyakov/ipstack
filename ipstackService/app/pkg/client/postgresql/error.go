package postgresql

import (
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
)

func parsePgError(err error) error {
	var pgErr *pgconn.PgError
	if errors.Is(err, pgErr) {
		pgErr = err.(*pgconn.PgError)
		return fmt.Errorf("database error. message:%s, detail:%s, where:%s, sqlstate:%s",
			pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.SQLState())
	}
	return err
}

func ErrCreateQuery(err error) error {
	return fmt.Errorf("failed to provide query due to %v", err)
}

func ErrExecuteQuery(err error) error {
	return fmt.Errorf("failed to execute query due to %v", err)
}

func ErrScanRow(err error) error {
	return fmt.Errorf("failed to scan row due to %v", err)
}
