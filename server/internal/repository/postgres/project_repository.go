package postgres

import (
	"context"
	"database/sql"

	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/repository"
	"github.com/google/uuid"
)

type projectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) repository.ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(ctx context.Context, project domain.Project) (domain.Project, error) {
	const query = `
		INSERT INTO project (id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, created_at, updated_at
	`
	var out domain.Project
	if err := r.db.QueryRowContext(
		ctx,
		query,
		project.ID,
		project.Name,
		project.Description,
	).Scan(
		&out.ID,
		&out.Name,
		&out.Description,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return domain.Project{}, MapError(err)
	}
	return out, nil
}

func (r *projectRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Project, error) {
	const query = `
		SELECT id, name, description, created_at, updated_at
		FROM project
		WHERE id = $1
	`
	var out domain.Project
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&out.ID,
		&out.Name,
		&out.Description,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return domain.Project{}, repository.ErrNotFound
		}
		return domain.Project{}, MapError(err)
	}
	return out, nil
}

func (r *projectRepository) List(ctx context.Context) ([]domain.Project, error) {
	const query = `
		SELECT id, name, description, created_at, updated_at
		FROM project
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, MapError(err)
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, MapError(err)
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, MapError(err)
	}
	return projects, nil
}

func (r *projectRepository) Update(ctx context.Context, id uuid.UUID, name string, description *string) (domain.Project, error) {
	const query = `
		UPDATE project
		SET name = $1, description = $2, updated_at = now()
		WHERE id = $3
		RETURNING id, name, description, created_at, updated_at
	`
	var out domain.Project
	if err := r.db.QueryRowContext(ctx, query, name, description, id).Scan(
		&out.ID,
		&out.Name,
		&out.Description,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return domain.Project{}, repository.ErrNotFound
		}
		return domain.Project{}, MapError(err)
	}
	return out, nil
}

func (r *projectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `
		DELETE FROM project
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return MapError(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return MapError(err)
	}
	if n == 0 {
		return repository.ErrNotFound
	}
	return nil
}
