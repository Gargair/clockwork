package http

import (
	"net/http"

	"github.com/Gargair/clockwork/server/internal/repository"
	"github.com/Gargair/clockwork/server/internal/service"
)

// apiErrorCode is the machine-readable error code used in ErrorResponse.
type apiErrorCode string

const (
	codeInvalidProjectName apiErrorCode = "invalid_project_name"
	codeInvalidParent      apiErrorCode = "invalid_parent"
	codeCrossProjectParent apiErrorCode = "cross_project_parent"
	codeCategoryCycle      apiErrorCode = "category_cycle"
	codeNoActiveTimer      apiErrorCode = "no_active_timer"
	codeInvalidJSON        apiErrorCode = "invalid_json"
	codeInvalidID          apiErrorCode = "invalid_id"
	codeInvalidTime        apiErrorCode = "invalid_time"
	codeNotFound           apiErrorCode = "not_found"
	codeInternal           apiErrorCode = "internal"
)

const (
	errInvalidJsonPayload              = "invalid JSON payload"
	errInvalidProjectId                = "invalid projectId"
	errInvalidCategoryId               = "invalid categoryId"
	errInvalidTime                     = "invalid time"
	errInvalidTimeRange                = "invalid time range"
	statusCodeFailedExpectationMessage = "expected %d, got %d"
)

// mapErrorToHTTP converts domain/service/repository errors into HTTP status and apiErrorCode.
func mapErrorToHTTP(err error) (status int, code apiErrorCode) {
	switch err {
	case service.ErrInvalidProjectName:
		return http.StatusBadRequest, codeInvalidProjectName
	case service.ErrInvalidParent:
		return http.StatusBadRequest, codeInvalidParent
	case service.ErrCrossProjectParent:
		return http.StatusBadRequest, codeCrossProjectParent
	case service.ErrCategoryCycle:
		return http.StatusConflict, codeCategoryCycle
	case service.ErrNoActiveTimer:
		return http.StatusConflict, codeNoActiveTimer
	case repository.ErrNotFound:
		return http.StatusNotFound, codeNotFound
	default:
		return http.StatusInternalServerError, codeInternal
	}
}

// writeMappedError writes an ErrorResponse computed from the given error.
// Message uses err.Error() so callers can provide user-facing details if desired.
func writeMappedError(w http.ResponseWriter, r *http.Request, err error) {
	status, code := mapErrorToHTTP(err)
	writeError(w, r, status, string(code), err.Error())
}
