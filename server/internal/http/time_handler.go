package http

import (
	"net/http"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/service"
)

// TimeHandler handles time tracking endpoints under /api/time.
type TimeHandler struct {
	svc    service.TimeTrackingService
	logger *slog.Logger
}

// NewTimeHandler constructs a TimeHandler.
func NewTimeHandler(svc service.TimeTrackingService, logger *slog.Logger) TimeHandler {
	return TimeHandler{svc: svc, logger: logger}
}

// RegisterRoutes mounts time routes under the provided router (expects base path /api/time).
func (h TimeHandler) RegisterRoutes(r chi.Router) {
	r.Post("/start", h.handleStart)
	r.Post("/stop", h.handleStop)
	r.Get("/active", h.handleActive)
	r.Get("/entries", h.handleEntries)
}

func (h TimeHandler) handleStart(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	var req TimeStartRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidJSON), errInvalidJsonPayload)
		h.logger.Warn("time_start_invalid_json", slog.String("request_id", reqID))
		return
	}
	catID, err := parseUUID(req.CategoryID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidCategoryId)
		h.logger.Warn("time_start_invalid_category", slog.String("request_id", reqID), slog.String("category_id", req.CategoryID))
		return
	}
	entry, err := h.svc.Start(r.Context(), catID)
	if err != nil {
		writeMappedError(w, r, err)
		h.logger.Error("time_start_error", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return
	}
	writeJSON(w, http.StatusCreated, timeEntryToResponse(entry))
	h.logger.Info("time_start_success", slog.String("request_id", reqID), slog.String("category_id", catID.String()))
}

func (h TimeHandler) handleStop(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	entry, err := h.svc.StopActive(r.Context())
	if err != nil {
		writeMappedError(w, r, err)
		h.logger.Warn("time_stop_no_active", slog.String("request_id", reqID))
		return
	}
	writeJSON(w, http.StatusOK, timeEntryToResponse(entry))
	h.logger.Info("time_stop_success", slog.String("request_id", reqID), slog.String("entry_id", entry.ID.String()))
}

func (h TimeHandler) handleActive(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	entry, err := h.svc.GetActive(r.Context())
	if err != nil {
		writeMappedError(w, r, err)
		h.logger.Error("time_active_error", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return
	}
	if entry == nil {
		// Return explicit null payload
		writeJSON(w, http.StatusOK, nil)
		h.logger.Info("time_active_none", slog.String("request_id", reqID))
		return
	}
	resp := timeEntryToResponse(*entry)
	writeJSON(w, http.StatusOK, resp)
	h.logger.Info("time_active_success", slog.String("request_id", reqID), slog.String("entry_id", resp.ID.String()))
}

func (h TimeHandler) handleEntries(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	q := r.URL.Query()
	catStr := q.Get("categoryId")
	if catStr == "" {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidCategoryId)
		h.logger.Warn("time_entries_missing_category", slog.String("request_id", reqID))
		return
	}
	catID, err := parseUUID(catStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidCategoryId)
		h.logger.Warn("time_entries_invalid_category", slog.String("request_id", reqID), slog.String("category_id", catStr))
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
			h.logger.Warn("time_entries_invalid_from", slog.String("request_id", reqID), slog.String("from", fromStr))
			return
		}
		from = t
		hasFrom = true
	}
	if toStr != "" {
		t, err := parseTimeRFC3339(toStr)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, string(codeInvalidTime), errInvalidTime)
			h.logger.Warn("time_entries_invalid_to", slog.String("request_id", reqID), slog.String("to", toStr))
			return
		}
		to = t
		hasTo = true
	}
	if hasFrom && hasTo && from.After(to) {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidTime), errInvalidTimeRange)
		h.logger.Warn("time_entries_invalid_range", slog.String("request_id", reqID))
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
		h.logger.Error("time_entries_error", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return
	}

	resp := make([]TimeEntryResponse, 0, len(entries))
	for _, e := range entries {
		resp = append(resp, timeEntryToResponse(e))
	}
	writeJSON(w, http.StatusOK, resp)
	h.logger.Info("time_entries_success", slog.String("request_id", reqID), slog.Int("count", len(resp)))
}

func timeEntryToResponse(e domain.TimeEntry) TimeEntryResponse {
	return TimeEntryResponse{
		ID:         e.ID,
		CategoryID: e.CategoryID,
		StartedAt:  e.StartedAt.UTC(),
		StoppedAt: func() *time.Time {
			if e.StoppedAt != nil {
				t := e.StoppedAt.UTC()
				return &t
			} else {
				return nil
			}
		}(),
		DurationSeconds: e.DurationSeconds,
		CreatedAt:       e.CreatedAt.UTC(),
		UpdatedAt:       e.UpdatedAt.UTC(),
	}
}
