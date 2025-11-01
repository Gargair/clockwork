package domain

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Category struct {
	ID               uuid.UUID
	ProjectID        uuid.UUID
	ParentCategoryID *uuid.UUID
	Name             string
	Description      *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type TimeEntry struct {
	ID              uuid.UUID
	CategoryID      uuid.UUID
	StartedAt       time.Time
	StoppedAt       *time.Time
	DurationSeconds *int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
}