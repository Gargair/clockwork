package http

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/service"
)

// ProjectHandler handles project endpoints.
type ProjectHandler struct {
	svc service.ProjectService
}

// NewProjectHandler constructs a ProjectHandler.
func NewProjectHandler(svc service.ProjectService) ProjectHandler {
	return ProjectHandler{svc: svc}
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
	var req ProjectCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidJSON), errInvalidJsonPayload)
		return
	}

	created, err := h.svc.Create(r.Context(), req.Name, req.Description)
	if err != nil {
		writeMappedError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, projectToResponse(created))
}

func (h ProjectHandler) handleList(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context())
	if err != nil {
		writeMappedError(w, r, err)
		return
	}
	resp := make([]ProjectResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, projectToResponse(it))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h ProjectHandler) handleGetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "projectId")
	id, err := parseUUID(idStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidProjectId)
		return
	}
	p, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeMappedError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, projectToResponse(p))
}

func (h ProjectHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "projectId")
	id, err := parseUUID(idStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidProjectId)
		return
	}

	var req ProjectUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidJSON), errInvalidJsonPayload)
		return
	}

	// For PATCH, require name to be provided (service requires non-empty name)
	if req.Name == nil {
		writeMappedError(w, r, service.ErrInvalidProjectName)
		return
	}
	name := strings.TrimSpace(*req.Name)
	desc := req.Description

	updated, err := h.svc.Update(r.Context(), id, name, desc)
	if err != nil {
		writeMappedError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, projectToResponse(updated))
}

func (h ProjectHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "projectId")
	id, err := parseUUID(idStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidProjectId)
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeMappedError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
