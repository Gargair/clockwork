package http

import (
	"net/http"
	"strings"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/service"
)

// ProjectHandler handles project endpoints.
type ProjectHandler struct {
	svc    service.ProjectService
	logger *slog.Logger
}

// NewProjectHandler constructs a ProjectHandler.
func NewProjectHandler(svc service.ProjectService, logger *slog.Logger) ProjectHandler {
	return ProjectHandler{svc: svc, logger: logger}
}

const projectIdRoute = "/{projectId}"

// RegisterRoutes mounts project routes under the provided router (expects base path to be set by caller).
func (h ProjectHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.handleCreate)
	r.Get("/", h.handleList)
	r.Get(projectIdRoute, h.handleGetByID)
	r.Patch(projectIdRoute, h.handleUpdate)
	r.Delete(projectIdRoute, h.handleDelete)
}

func (h ProjectHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	h.logger.Info("project_create_start", slog.String("request_id", reqID))
	var req ProjectCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidJSON), errInvalidJsonPayload)
		h.logger.Warn("project_create_invalid_json", slog.String("request_id", reqID))
		return
	}

	created, err := h.svc.Create(r.Context(), req.Name, req.Description)
	if err != nil {
		writeMappedError(w, r, err)
		h.logger.Error("project_create_error", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return
	}
	writeJSON(w, http.StatusCreated, projectToResponse(created))
	h.logger.Info("project_create_success", slog.String("request_id", reqID), slog.String("project_id", created.ID.String()))
}

func (h ProjectHandler) handleList(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	h.logger.Info("project_list_start", slog.String("request_id", reqID))
	items, err := h.svc.List(r.Context())
	if err != nil {
		writeMappedError(w, r, err)
		h.logger.Error("project_list_error", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return
	}
	resp := make([]ProjectResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, projectToResponse(it))
	}
	writeJSON(w, http.StatusOK, resp)
	h.logger.Info("project_list_success", slog.String("request_id", reqID), slog.Int("count", len(resp)))
}

func (h ProjectHandler) handleGetByID(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	idStr := chi.URLParam(r, "projectId")
	id, err := parseUUID(idStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidProjectId)
		h.logger.Warn("project_get_invalid_id", slog.String("request_id", reqID), slog.String("project_id", idStr))
		return
	}
	p, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeMappedError(w, r, err)
		h.logger.Error("project_get_error", slog.String("request_id", reqID), slog.String("error", err.Error()), slog.String("project_id", id.String()))
		return
	}
	writeJSON(w, http.StatusOK, projectToResponse(p))
	h.logger.Info("project_get_success", slog.String("request_id", reqID), slog.String("project_id", id.String()))
}

func (h ProjectHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	idStr := chi.URLParam(r, "projectId")
	id, err := parseUUID(idStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidProjectId)
		h.logger.Warn("project_update_invalid_id", slog.String("request_id", reqID), slog.String("project_id", idStr))
		return
	}

	var req ProjectUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidJSON), errInvalidJsonPayload)
		h.logger.Warn("project_update_invalid_json", slog.String("request_id", reqID))
		return
	}

	// For PATCH, require name to be provided (service requires non-empty name)
	if req.Name == nil {
		writeMappedError(w, r, service.ErrInvalidProjectName)
		h.logger.Warn("project_update_missing_name", slog.String("request_id", reqID), slog.String("project_id", id.String()))
		return
	}
	name := strings.TrimSpace(*req.Name)
	desc := req.Description

	updated, err := h.svc.Update(r.Context(), id, name, desc)
	if err != nil {
		writeMappedError(w, r, err)
		h.logger.Error("project_update_error", slog.String("request_id", reqID), slog.String("error", err.Error()), slog.String("project_id", id.String()))
		return
	}
	writeJSON(w, http.StatusOK, projectToResponse(updated))
	h.logger.Info("project_update_success", slog.String("request_id", reqID), slog.String("project_id", id.String()))
}

func (h ProjectHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	idStr := chi.URLParam(r, "projectId")
	id, err := parseUUID(idStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidProjectId)
		h.logger.Warn("project_delete_invalid_id", slog.String("request_id", reqID), slog.String("project_id", idStr))
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeMappedError(w, r, err)
		h.logger.Error("project_delete_error", slog.String("request_id", reqID), slog.String("error", err.Error()), slog.String("project_id", id.String()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
	h.logger.Info("project_delete_success", slog.String("request_id", reqID), slog.String("project_id", id.String()))
}

func projectToResponse(p domain.Project) ProjectResponse {
	return ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
