package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/service"
)

// TimeHandler handles time tracking endpoints under /api/time.
type TimeHandler struct {
	svc service.TimeTrackingService
}

// NewTimeHandler constructs a TimeHandler.
func NewTimeHandler(svc service.TimeTrackingService) TimeHandler { return TimeHandler{svc: svc} }

// RegisterRoutes mounts time routes under the provided router (expects base path /api/time).
func (h TimeHandler) RegisterRoutes(r chi.Router) {
	r.Post("/start", h.handleStart)
	r.Post("/stop", h.handleStop)
	r.Get("/active", h.handleActive)
	r.Get("/entries", h.handleEntries)
}

func (h TimeHandler) handleStart(w http.ResponseWriter, r *http.Request) {
	var req TimeStartRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidJSON), errInvalidJsonPayload)
		return
	}
	catID, err := parseUUID(req.CategoryID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidCategoryId)
		return
	}
	entry, err := h.svc.Start(r.Context(), catID)
	if err != nil {
		writeMappedError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, timeEntryToResponse(entry))
}

func (h TimeHandler) handleStop(w http.ResponseWriter, r *http.Request) {
	entry, err := h.svc.StopActive(r.Context())
	if err != nil {
		writeMappedError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, timeEntryToResponse(entry))
}

func (h TimeHandler) handleActive(w http.ResponseWriter, r *http.Request) {
	entry, err := h.svc.GetActive(r.Context())
	if err != nil {
		writeMappedError(w, r, err)
		return
	}
	if entry == nil {
		// Return explicit null payload
		writeJSON(w, http.StatusOK, nil)
		return
	}
	resp := timeEntryToResponse(*entry)
	writeJSON(w, http.StatusOK, resp)
}

func (h TimeHandler) handleEntries(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	catStr := q.Get("categoryId")
	if catStr == "" {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidCategoryId)
		return
	}
	catID, err := parseUUID(catStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidCategoryId)
		return
	}

	fromStr := q.Get("from")
	toStr := q.Get("to")
	var (
		from    time.Time
		to      time.Time
		hasFrom bool
		hasTo   bool
	)
	if fromStr != "" {
		t, err := parseTimeRFC3339(fromStr)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, string(codeInvalidTime), errInvalidTime)
			return
		}
		from = t
		hasFrom = true
	}
	if toStr != "" {
		t, err := parseTimeRFC3339(toStr)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, string(codeInvalidTime), errInvalidTime)
			return
		}
		to = t
		hasTo = true
	}
	if hasFrom && hasTo && from.After(to) {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidTime), errInvalidTimeRange)
		return
	}

	var entries []domain.TimeEntry
	if hasFrom && hasTo {
		entries, err = h.svc.ListByCategoryAndRange(r.Context(), catID, from, to)
	} else {
		entries, err = h.svc.ListByCategory(r.Context(), catID)
	}
	if err != nil {
		writeMappedError(w, r, err)
		return
	}

	resp := make([]TimeEntryResponse, 0, len(entries))
	for _, e := range entries {
		resp = append(resp, timeEntryToResponse(e))
	}
	writeJSON(w, http.StatusOK, resp)
}

func timeEntryToResponse(e domain.TimeEntry) TimeEntryResponse {
	return TimeEntryResponse{
		ID:              e.ID,
		CategoryID:      e.CategoryID,
		StartedAt:       e.StartedAt,
		StoppedAt:       e.StoppedAt,
		DurationSeconds: e.DurationSeconds,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}
