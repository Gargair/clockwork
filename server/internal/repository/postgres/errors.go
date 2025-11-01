package postgres

import (
	"errors"

	"github.com/Gargair/clockwork/server/internal/repository"
	"github.com/jackc/pgconn"
)

// MapError translates low-level Postgres errors into repository-level errors where appropriate.
// Unknown errors are returned unchanged.
func MapError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return repository.ErrDuplicate
		case "23503": // foreign_key_violation
			return repository.ErrForeignKeyViolation
		}
	}
	return err
}
