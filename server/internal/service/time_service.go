package service

import (
	"context"
	"time"

	"github.com/Gargair/clockwork/server/internal/clock"
	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/repository"
	"github.com/google/uuid"
)

type timeTrackingService struct {
	repo         repository.TimeEntryRepository
	categoryRepo repository.CategoryRepository
	clk          clock.Clock
}

func (s *timeTrackingService) Start(ctx context.Context, categoryID uuid.UUID) (domain.TimeEntry, error) {
	// Ensure category exists
	if _, err := s.categoryRepo.GetByID(ctx, categoryID); err != nil {
		return domain.TimeEntry{}, err
	}

	now := s.clk.Now()

	// If an active entry exists, stop it using the same timestamp and computed duration
	active, err := s.repo.FindActive(ctx)
	if err != nil {
		return domain.TimeEntry{}, err
	}
	if active != nil {
		durationSeconds := int32(now.Sub(active.StartedAt).Seconds())
		if durationSeconds < 0 {
			durationSeconds = 0
		}
		if _, err := s.repo.Stop(ctx, active.ID, now, &durationSeconds); err != nil {
			return domain.TimeEntry{}, err
		}
	}

	// Create new active entry
	entry := domain.TimeEntry{
		ID:              uuid.New(),
		CategoryID:      categoryID,
		StartedAt:       now,
		StoppedAt:       nil,
		DurationSeconds: nil,
	}
	return s.repo.Create(ctx, entry)
}

func (s *timeTrackingService) StopActive(ctx context.Context) (domain.TimeEntry, error) {
	active, err := s.repo.FindActive(ctx)
	if err != nil {
		return domain.TimeEntry{}, err
	}
	if active == nil {
		return domain.TimeEntry{}, ErrNoActiveTimer
	}
	now := s.clk.Now()
	durationSeconds := int32(now.Sub(active.StartedAt).Seconds())
	if durationSeconds < 0 {
		durationSeconds = 0
	}
	return s.repo.Stop(ctx, active.ID, now, &durationSeconds)
}

func (s *timeTrackingService) GetActive(ctx context.Context) (*domain.TimeEntry, error) {
	return s.repo.FindActive(ctx)
}

func (s *timeTrackingService) ListByCategory(ctx context.Context, categoryID uuid.UUID) ([]domain.TimeEntry, error) {
	return s.repo.ListByCategory(ctx, categoryID)
}

func (s *timeTrackingService) ListByCategoryAndRange(ctx context.Context, categoryID uuid.UUID, start time.Time, end time.Time) ([]domain.TimeEntry, error) {
	return s.repo.ListByCategoryAndRange(ctx, categoryID, start, end)
}

var _ TimeTrackingService = (*timeTrackingService)(nil)
