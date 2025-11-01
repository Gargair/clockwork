package postgres

import (
	"context"
	"database/sql"

	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/repository"
	"github.com/google/uuid"
)

type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) repository.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category domain.Category) (domain.Category, error) {
	const query = `
		INSERT INTO category (id, project_id, parent_category_id, name, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, project_id, parent_category_id, name, description, created_at, updated_at
	`
	var out domain.Category
	if err := r.db.QueryRowContext(
		ctx,
		query,
		category.ID,
		category.ProjectID,
		category.ParentCategoryID,
		category.Name,
		category.Description,
	).Scan(
		&out.ID,
		&out.ProjectID,
		&out.ParentCategoryID,
		&out.Name,
		&out.Description,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return domain.Category{}, MapError(err)
	}
	return out, nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Category, error) {
	const query = `
		SELECT id, project_id, parent_category_id, name, description, created_at, updated_at
		FROM category
		WHERE id = $1
	`
	var out domain.Category
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&out.ID,
		&out.ProjectID,
		&out.ParentCategoryID,
		&out.Name,
		&out.Description,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return domain.Category{}, repository.ErrNotFound
		}
		return domain.Category{}, MapError(err)
	}
	return out, nil
}

func (r *categoryRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Category, error) {
	const query = `
		SELECT id, project_id, parent_category_id, name, description, created_at, updated_at
		FROM category
		WHERE project_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, MapError(err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.ParentCategoryID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, MapError(err)
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, MapError(err)
	}
	return categories, nil
}

func (r *categoryRepository) ListChildren(ctx context.Context, parentID uuid.UUID) ([]domain.Category, error) {
	const query = `
		SELECT id, project_id, parent_category_id, name, description, created_at, updated_at
		FROM category
		WHERE parent_category_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, MapError(err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.ParentCategoryID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, MapError(err)
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, MapError(err)
	}
	return categories, nil
}

func (r *categoryRepository) Update(ctx context.Context, id uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error) {
	const query = `
		UPDATE category
		SET name = $1, description = $2, parent_category_id = $3, updated_at = now()
		WHERE id = $4
		RETURNING id, project_id, parent_category_id, name, description, created_at, updated_at
	`
	var out domain.Category
	if err := r.db.QueryRowContext(ctx, query, name, description, parentCategoryID, id).Scan(
		&out.ID,
		&out.ProjectID,
		&out.ParentCategoryID,
		&out.Name,
		&out.Description,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return domain.Category{}, repository.ErrNotFound
		}
		return domain.Category{}, MapError(err)
	}
	return out, nil
}

func (r *categoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `
		DELETE FROM category
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
