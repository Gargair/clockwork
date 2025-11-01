package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/repository"
	"github.com/google/uuid"
)

type timeEntryRepository struct {
	db *sql.DB
}

func NewTimeEntryRepository(db *sql.DB) repository.TimeEntryRepository {
	return &timeEntryRepository{db: db}
}

func (r *timeEntryRepository) Create(ctx context.Context, entry domain.TimeEntry) (domain.TimeEntry, error) {
	const query = `
		INSERT INTO time_entry (id, category_id, started_at, stopped_at, duration_seconds)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, category_id, started_at, stopped_at, duration_seconds, created_at, updated_at
	`
	var out domain.TimeEntry
	if err := r.db.QueryRowContext(
		ctx,
		query,
		entry.ID,
		entry.CategoryID,
		entry.StartedAt,
		entry.StoppedAt,
		entry.DurationSeconds,
	).Scan(
		&out.ID,
		&out.CategoryID,
		&out.StartedAt,
		&out.StoppedAt,
		&out.DurationSeconds,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return domain.TimeEntry{}, MapError(err)
	}
	return out, nil
}

func (r *timeEntryRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.TimeEntry, error) {
	const query = `
		SELECT id, category_id, started_at, stopped_at, duration_seconds, created_at, updated_at
		FROM time_entry
		WHERE id = $1
	`
	var out domain.TimeEntry
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&out.ID,
		&out.CategoryID,
		&out.StartedAt,
		&out.StoppedAt,
		&out.DurationSeconds,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return domain.TimeEntry{}, repository.ErrNotFound
		}
		return domain.TimeEntry{}, MapError(err)
	}
	return out, nil
}

func (r *timeEntryRepository) ListByCategory(ctx context.Context, categoryID uuid.UUID) ([]domain.TimeEntry, error) {
	const query = `
		SELECT id, category_id, started_at, stopped_at, duration_seconds, created_at, updated_at
		FROM time_entry
		WHERE category_id = $1
		ORDER BY started_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, categoryID)
	if err != nil {
		return nil, MapError(err)
	}
	defer rows.Close()

	var entries []domain.TimeEntry
	for rows.Next() {
		var e domain.TimeEntry
		if err := rows.Scan(&e.ID, &e.CategoryID, &e.StartedAt, &e.StoppedAt, &e.DurationSeconds, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, MapError(err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, MapError(err)
	}
	return entries, nil
}

func (r *timeEntryRepository) ListByCategoryAndRange(ctx context.Context, categoryID uuid.UUID, start time.Time, end time.Time) ([]domain.TimeEntry, error) {
	const query = `
		SELECT id, category_id, started_at, stopped_at, duration_seconds, created_at, updated_at
		FROM time_entry
		WHERE category_id = $1 AND started_at >= $2 AND started_at <= $3
		ORDER BY started_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, categoryID, start, end)
	if err != nil {
		return nil, MapError(err)
	}
	defer rows.Close()

	var entries []domain.TimeEntry
	for rows.Next() {
		var e domain.TimeEntry
		if err := rows.Scan(&e.ID, &e.CategoryID, &e.StartedAt, &e.StoppedAt, &e.DurationSeconds, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, MapError(err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, MapError(err)
	}
	return entries, nil
}

func (r *timeEntryRepository) FindActive(ctx context.Context) (*domain.TimeEntry, error) {
	const query = `
		SELECT id, category_id, started_at, stopped_at, duration_seconds, created_at, updated_at
		FROM time_entry
		WHERE stopped_at IS NULL
		ORDER BY started_at DESC
		LIMIT 1
	`
	var out domain.TimeEntry
	if err := r.db.QueryRowContext(ctx, query).Scan(
		&out.ID,
		&out.CategoryID,
		&out.StartedAt,
		&out.StoppedAt,
		&out.DurationSeconds,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, MapError(err)
	}
	return &out, nil
}

func (r *timeEntryRepository) Stop(ctx context.Context, id uuid.UUID, stoppedAt time.Time, durationSeconds *int32) (domain.TimeEntry, error) {
	const query = `
		UPDATE time_entry
		SET stopped_at = $2, duration_seconds = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, category_id, started_at, stopped_at, duration_seconds, created_at, updated_at
	`
	var out domain.TimeEntry
	if err := r.db.QueryRowContext(ctx, query, id, stoppedAt, durationSeconds).Scan(
		&out.ID,
		&out.CategoryID,
		&out.StartedAt,
		&out.StoppedAt,
		&out.DurationSeconds,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return domain.TimeEntry{}, repository.ErrNotFound
		}
		return domain.TimeEntry{}, MapError(err)
	}
	return out, nil
}
