package http

import (
	"time"

	"github.com/google/uuid"
)

// ProjectCreateRequest represents the payload to create a project.
type ProjectCreateRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// ProjectUpdateRequest represents the payload to update a project.
// Fields are optional to support PATCH semantics; handlers will validate.
type ProjectUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// ProjectResponse is the API response shape for a project.
type ProjectResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CategoryCreateRequest represents the payload to create a category.
type CategoryCreateRequest struct {
	Name             string  `json:"name"`
	Description      *string `json:"description,omitempty"`
	ParentCategoryID *string `json:"parentCategoryId"`
}

// CategoryUpdateRequest represents the payload to update a category.
// Fields are optional to support PATCH semantics; handlers will validate.
type CategoryUpdateRequest struct {
	Name             *string `json:"name,omitempty"`
	Description      *string `json:"description,omitempty"`
	ParentCategoryID *string `json:"parentCategoryId"`
}

// CategoryResponse is the API response shape for a category.
type CategoryResponse struct {
	ID               uuid.UUID  `json:"id"`
	ProjectID        uuid.UUID  `json:"projectId"`
	ParentCategoryID *uuid.UUID `json:"parentCategoryId"`
	Name             string     `json:"name"`
	Description      *string    `json:"description,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// TimeStartRequest represents the payload to start a timer.
type TimeStartRequest struct {
	CategoryID string `json:"categoryId"`
}

// TimeEntryResponse is the API response shape for a time entry.
type TimeEntryResponse struct {
	ID              uuid.UUID  `json:"id"`
	CategoryID      uuid.UUID  `json:"categoryId"`
	StartedAt       time.Time  `json:"startedAt"`
	StoppedAt       *time.Time `json:"stoppedAt"`
	DurationSeconds *int32     `json:"durationSeconds"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// ActiveTimerResponse represents the response of the active timer endpoint when an entry exists.
// When no active timer exists, the endpoint should return a JSON null.
type ActiveTimerResponse = TimeEntryResponse

// ErrorResponse is the standard error envelope for API errors.
type ErrorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}
