package service

import (
	"context"

	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/repository"
	"github.com/google/uuid"
)

type categoryService struct {
	repo repository.CategoryRepository
}

func (s *categoryService) Create(ctx context.Context, projectID uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error) {
	if parentCategoryID != nil {
		parent, err := s.repo.GetByID(ctx, *parentCategoryID)
		if err != nil {
			if err == repository.ErrNotFound {
				return domain.Category{}, ErrInvalidParent
			}
			return domain.Category{}, err
		}
		if parent.ProjectID != projectID {
			return domain.Category{}, ErrCrossProjectParent
		}
	}

	c := domain.Category{
		ID:               uuid.New(),
		ProjectID:        projectID,
		ParentCategoryID: parentCategoryID,
		Name:             name,
		Description:      description,
	}
	return s.repo.Create(ctx, c)
}

func (s *categoryService) Update(ctx context.Context, id uuid.UUID, name string, description *string, parentCategoryID *uuid.UUID) (domain.Category, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Category{}, err
	}

	if parentCategoryID != nil {
		// Parent must exist and belong to same project
		parent, err := s.repo.GetByID(ctx, *parentCategoryID)
		if err != nil {
			if err == repository.ErrNotFound {
				return domain.Category{}, ErrInvalidParent
			}
			return domain.Category{}, err
		}
		if parent.ProjectID != current.ProjectID {
			return domain.Category{}, ErrCrossProjectParent
		}
		// No cycles: parent cannot be self or any descendant of self
		if *parentCategoryID == id {
			return domain.Category{}, ErrCategoryCycle
		}
		isDesc, err := s.isDescendant(ctx, id, *parentCategoryID)
		if err != nil {
			return domain.Category{}, err
		}
		if isDesc {
			return domain.Category{}, ErrCategoryCycle
		}
	}

	return s.repo.Update(ctx, id, name, description, parentCategoryID)
}

func (s *categoryService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *categoryService) GetByID(ctx context.Context, id uuid.UUID) (domain.Category, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *categoryService) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Category, error) {
	return s.repo.ListByProject(ctx, projectID)
}

func (s *categoryService) ListChildren(ctx context.Context, parentID uuid.UUID) ([]domain.Category, error) {
	return s.repo.ListChildren(ctx, parentID)
}

// isDescendant checks whether candidateID is a descendant of rootID using BFS over ListChildren.
func (s *categoryService) isDescendant(ctx context.Context, rootID uuid.UUID, candidateID uuid.UUID) (bool, error) {
	visited := map[uuid.UUID]struct{}{}
	queue := []uuid.UUID{rootID}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if _, ok := visited[cur]; ok {
			continue
		}
		visited[cur] = struct{}{}
		children, err := s.repo.ListChildren(ctx, cur)
		if err != nil {
			return false, err
		}
		for _, ch := range children {
			if ch.ID == candidateID {
				return true, nil
			}
			queue = append(queue, ch.ID)
		}
	}
	return false, nil
}

var _ CategoryService = (*categoryService)(nil)
