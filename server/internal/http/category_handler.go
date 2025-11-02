package http

import (
	"net/http"
	"strings"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/Gargair/clockwork/server/internal/domain"
	"github.com/Gargair/clockwork/server/internal/service"
)

// CategoryHandler handles category endpoints under a project.
type CategoryHandler struct {
	svc    service.CategoryService
	logger *slog.Logger
}

// NewCategoryHandler constructs a CategoryHandler.
func NewCategoryHandler(svc service.CategoryService, logger *slog.Logger) CategoryHandler {
	return CategoryHandler{svc: svc, logger: logger}
}

const (
	projectIdParam  = "projectId"
	categoryIdRoute = "/{categoryId}"
	categoryIdParam = "categoryId"
)

// RegisterRoutes mounts category routes on the provided router (expects base path to be /api/projects/{projectId}/categories).
func (h CategoryHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.handleCreate)
	r.Get("/", h.handleList)
	r.Get(categoryIdRoute, h.handleGetByID)
	r.Patch(categoryIdRoute, h.handleUpdate)
	r.Delete(categoryIdRoute, h.handleDelete)
}

func (h CategoryHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	projID, ok := h.parseProjectID(w, r)
	if !ok {
		return
	}
	var req CategoryCreateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidJSON), errInvalidJsonPayload)
		h.logger.Warn("category_create_invalid_json", slog.String("request_id", reqID))
		return
	}
	parentUUID, err := parseOptionalUUID(req.ParentCategoryID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidProjectId)
		h.logger.Warn("category_create_invalid_parent", slog.String("request_id", reqID))
		return
	}

	created, err := h.svc.Create(r.Context(), projID, strings.TrimSpace(req.Name), req.Description, parentUUID)
	if err != nil {
		writeMappedError(w, r, err)
		h.logger.Error("category_create_error", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return
	}
	writeJSON(w, http.StatusCreated, categoryToResponse(created))
	h.logger.Info("category_create_success", slog.String("request_id", reqID), slog.String("category_id", created.ID.String()))
}

func (h CategoryHandler) handleList(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	projID, ok := h.parseProjectID(w, r)
	if !ok {
		return
	}
	items, err := h.svc.ListByProject(r.Context(), projID)
	if err != nil {
		writeMappedError(w, r, err)
		h.logger.Error("category_list_error", slog.String("request_id", reqID), slog.String("error", err.Error()))
		return
	}
	resp := make([]CategoryResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, categoryToResponse(it))
	}
	writeJSON(w, http.StatusOK, resp)
	h.logger.Info("category_list_success", slog.String("request_id", reqID), slog.Int("count", len(resp)))
}

func (h CategoryHandler) handleGetByID(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	if _, ok := h.parseProjectID(w, r); !ok {
		return
	}
	idStr := chi.URLParam(r, "categoryId")
	id, err := parseUUID(idStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidProjectId)
		h.logger.Warn("category_get_invalid_id", slog.String("request_id", reqID), slog.String("category_id", idStr))
		return
	}
	c, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeMappedError(w, r, err)
		h.logger.Error("category_get_error", slog.String("request_id", reqID), slog.String("error", err.Error()), slog.String("category_id", id.String()))
		return
	}
	writeJSON(w, http.StatusOK, categoryToResponse(c))
	h.logger.Info("category_get_success", slog.String("request_id", reqID), slog.String("category_id", id.String()))
}

func (h CategoryHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	if _, ok := h.parseProjectID(w, r); !ok {
		return
	}
	idStr := chi.URLParam(r, "categoryId")
	id, err := parseUUID(idStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidProjectId)
		h.logger.Warn("category_update_invalid_id", slog.String("request_id", reqID), slog.String("category_id", idStr))
		return
	}

	var req CategoryUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidJSON), errInvalidJsonPayload)
		h.logger.Warn("category_update_invalid_json", slog.String("request_id", reqID))
		return
	}
	if req.Name == nil {
		writeMappedError(w, r, service.ErrInvalidParent) // force 400; name required by service.Update
		h.logger.Warn("category_update_missing_name", slog.String("request_id", reqID), slog.String("category_id", id.String()))
		return
	}
	name := strings.TrimSpace(*req.Name)
	parentUUID, err := parseOptionalUUID(req.ParentCategoryID)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidProjectId)
		h.logger.Warn("category_update_invalid_parent", slog.String("request_id", reqID))
		return
	}

	updated, err := h.svc.Update(r.Context(), id, name, req.Description, parentUUID)
	if err != nil {
		writeMappedError(w, r, err)
		h.logger.Error("category_update_error", slog.String("request_id", reqID), slog.String("error", err.Error()), slog.String("category_id", id.String()))
		return
	}
	writeJSON(w, http.StatusOK, categoryToResponse(updated))
	h.logger.Info("category_update_success", slog.String("request_id", reqID), slog.String("category_id", id.String()))
}

func (h CategoryHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	if _, ok := h.parseProjectID(w, r); !ok {
		return
	}
	idStr := chi.URLParam(r, "categoryId")
	id, err := parseUUID(idStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidProjectId)
		h.logger.Warn("category_delete_invalid_id", slog.String("request_id", reqID), slog.String("category_id", idStr))
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeMappedError(w, r, err)
		h.logger.Error("category_delete_error", slog.String("request_id", reqID), slog.String("error", err.Error()), slog.String("category_id", id.String()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
	h.logger.Info("category_delete_success", slog.String("request_id", reqID), slog.String("category_id", id.String()))
}

func (h CategoryHandler) parseProjectID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr := chi.URLParam(r, projectIdParam)
	id, err := parseUUID(idStr)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, string(codeInvalidID), errInvalidProjectId)
		return uuid.Nil, false
	}
	return id, true
}

func parseOptionalUUID(s *string) (*uuid.UUID, error) {
	if s == nil {
		return nil, nil
	}
	u, err := parseUUID(*s)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func categoryToResponse(c domain.Category) CategoryResponse {
	return CategoryResponse{
		ID:               c.ID,
		ProjectID:        c.ProjectID,
		ParentCategoryID: c.ParentCategoryID,
		Name:             c.Name,
		Description:      c.Description,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
	}
}
