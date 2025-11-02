package service

import (
	"context"
	"errors"
	"strings"

	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/repository"
	"github.com/google/uuid"
)

type projectService struct {
	repo repository.ProjectRepository
}

func (s *projectService) Create(ctx context.Context, name string, description *string) (domain.Project, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return domain.Project{}, errors.New("service: project name cannot be empty")
	}
	p := domain.Project{
		ID:          uuid.New(),
		Name:        trimmed,
		Description: description,
	}
	return s.repo.Create(ctx, p)
}

func (s *projectService) Update(ctx context.Context, id uuid.UUID, name string, description *string) (domain.Project, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return domain.Project{}, errors.New("service: project name cannot be empty")
	}
	return s.repo.Update(ctx, id, trimmed, description)
}

func (s *projectService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *projectService) GetByID(ctx context.Context, id uuid.UUID) (domain.Project, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *projectService) List(ctx context.Context) ([]domain.Project, error) {
	return s.repo.List(ctx)
}

var _ ProjectService = (*projectService)(nil)
