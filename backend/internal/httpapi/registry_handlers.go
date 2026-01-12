package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/registries"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RegistriesHandler struct {
	service *registries.Service
}

func NewRegistriesHandler(service *registries.Service) *RegistriesHandler {
	return &RegistriesHandler{service: service}
}

type registryRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	RegistryURL string `json:"registry_url"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

type registryResponse struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	RegistryURL string    `json:"registry_url"`
	Username    *string   `json:"username,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *RegistriesHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseUUIDParam(r, "projectId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	items, err := h.service.ListRegistries(r.Context(), user.ID, projectID)
	if err != nil {
		if errors.Is(err, registries.ErrProjectNotFound) {
			writeError(w, http.StatusNotFound, "project not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to list registries")
		return
	}

	resp := make([]registryResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toRegistryResponse(item))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *RegistriesHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseUUIDParam(r, "projectId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var req registryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	registry, err := h.service.CreateRegistry(r.Context(), user.ID, projectID, registries.CreateRegistryInput{
		Name:        req.Name,
		Type:        req.Type,
		RegistryURL: req.RegistryURL,
		Username:    req.Username,
		Password:    req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, registries.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, registries.ErrNotOwner):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, registries.ErrRegistryNameConflict):
			writeError(w, http.StatusConflict, "registry with this name already exists")
		case errors.Is(err, registries.ErrInvalidRegistryName),
			errors.Is(err, registries.ErrInvalidRegistryType),
			errors.Is(err, registries.ErrInvalidRegistryURL):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to create registry")
		}
		return
	}

	writeJSON(w, http.StatusCreated, toRegistryResponse(registry))
}

func (h *RegistriesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := parseUUIDParam(r, "projectId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	registryID, err := parseUUIDParam(r, "registryId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid registry id")
		return
	}

	if err := h.service.DeleteRegistry(r.Context(), user.ID, projectID, registryID); err != nil {
		switch {
		case errors.Is(err, registries.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, registries.ErrNotOwner):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, registries.ErrRegistryNotFound):
			writeError(w, http.StatusNotFound, "registry not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete registry")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	value := chi.URLParam(r, name)
	return uuid.Parse(value)
}

func toRegistryResponse(registry registries.Registry) registryResponse {
	return registryResponse{
		ID:          registry.ID.String(),
		ProjectID:   registry.ProjectID.String(),
		Name:        registry.Name,
		Type:        registry.Type,
		RegistryURL: registry.RegistryURL,
		Username:    registry.Username,
		CreatedAt:   registry.CreatedAt,
		UpdatedAt:   registry.UpdatedAt,
	}
}
