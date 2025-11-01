package postgres

import (
	"errors"
	"strings"

	"github.com/Gargair/clockwork/server/internal/repository"
	"github.com/jackc/pgconn"
)

// MapError translates low-level Postgres errors into repository-level errors where appropriate.
// Unknown errors are returned unchanged.
func MapError(err error) error {
	if err == nil {
		return nil
	}
	// Try pointer form first (the common case with pgx)
	var pgErrPtr *pgconn.PgError
	if errors.As(err, &pgErrPtr) {
		switch pgErrPtr.Code {
		case "23505": // unique_violation
			return repository.ErrDuplicate
		case "23503": // foreign_key_violation
			return repository.ErrForeignKeyViolation
		}
	}
	// Fallback: match common SQLSTATE codes in the error string
	msg := err.Error()
	if strings.Contains(msg, "SQLSTATE 23505") { // unique_violation
		return repository.ErrDuplicate
	}
	if strings.Contains(msg, "SQLSTATE 23503") { // foreign_key_violation
		return repository.ErrForeignKeyViolation
	}
	return err
}
