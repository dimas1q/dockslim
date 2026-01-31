package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/budgets"
	"github.com/dimas1q/dockslim/backend/internal/projects"
	"github.com/google/uuid"
)

type budgetRepoStub struct {
	list     []budgets.Budget
	conflict bool
	calls    int
}

func (r *budgetRepoStub) ListBudgetsByProject(ctx context.Context, projectID uuid.UUID) ([]budgets.Budget, error) {
	return r.list, nil
}

func (r *budgetRepoStub) UpsertDefaultBudget(ctx context.Context, projectID uuid.UUID, thresholds budgets.ResolvedBudget) (budgets.Budget, error) {
	return budgets.Budget{ID: uuid.New(), ProjectID: projectID, WarnDeltaBytes: thresholds.WarnDeltaBytes}, nil
}

func (r *budgetRepoStub) CreateBudgetOverride(ctx context.Context, projectID uuid.UUID, image string, thresholds budgets.ResolvedBudget) (budgets.Budget, error) {
	r.calls++
	if r.conflict && r.calls > 1 {
		return budgets.Budget{}, budgets.ErrBudgetConflict
	}
	return budgets.Budget{ID: uuid.New(), ProjectID: projectID, Image: &image, WarnDeltaBytes: thresholds.WarnDeltaBytes}, nil
}

func (r *budgetRepoStub) UpdateBudget(ctx context.Context, budgetID, projectID uuid.UUID, image *string, thresholds budgets.ResolvedBudget) (budgets.Budget, error) {
	return budgets.Budget{ID: budgetID, ProjectID: projectID, Image: image, WarnDeltaBytes: thresholds.WarnDeltaBytes}, nil
}

func (r *budgetRepoStub) DeleteBudget(ctx context.Context, budgetID, projectID uuid.UUID) error {
	return nil
}

func (r *budgetRepoStub) ResolveBudgetForImage(ctx context.Context, projectID uuid.UUID, image string) (*budgets.ResolvedBudget, error) {
	return nil, nil
}

type budgetMemberStub struct {
	role string
	err  error
}

func (m *budgetMemberStub) GetMemberRole(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.role, nil
}

func TestBudgetsListMemberAllowed(t *testing.T) {
	repo := &budgetRepoStub{list: []budgets.Budget{}}
	members := &budgetMemberStub{role: "member"}
	svc := budgets.NewService(repo, members)
	handler := NewBudgetsHandler(svc)

	projectID := uuid.New()
	user := auth.User{ID: uuid.New()}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+projectID.String()+"/budgets", nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "projectId", projectID.String())

	rr := httptest.NewRecorder()
	handler.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestBudgetsUpsertOwnerOnly(t *testing.T) {
	repo := &budgetRepoStub{}
	members := &budgetMemberStub{role: "member"}
	svc := budgets.NewService(repo, members)
	handler := NewBudgetsHandler(svc)

	projectID := uuid.New()
	user := auth.User{ID: uuid.New()}
	payload := map[string]int64{"warn_delta_mb": 10}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/"+projectID.String()+"/budgets/default", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "projectId", projectID.String())

	rr := httptest.NewRecorder()
	handler.UpsertDefault(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestBudgetsCreateOverrideValidation(t *testing.T) {
	repo := &budgetRepoStub{}
	members := &budgetMemberStub{role: projects.RoleOwner}
	svc := budgets.NewService(repo, members)
	handler := NewBudgetsHandler(svc)

	projectID := uuid.New()
	user := auth.User{ID: uuid.New()}
	payload := map[string]interface{}{"image": "", "warn_delta_mb": 5}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID.String()+"/budgets/overrides", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "projectId", projectID.String())

	rr := httptest.NewRecorder()
	handler.CreateOverride(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestBudgetsCreateOverrideConflict(t *testing.T) {
	repo := &budgetRepoStub{conflict: true}
	members := &budgetMemberStub{role: projects.RoleOwner}
	svc := budgets.NewService(repo, members)
	handler := NewBudgetsHandler(svc)

	projectID := uuid.New()
	user := auth.User{ID: uuid.New()}
	payload := map[string]interface{}{"image": "company/app", "warn_delta_mb": 5}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID.String()+"/budgets/overrides", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = withURLParamAnalysis(req, "projectId", projectID.String())

	rr := httptest.NewRecorder()
	handler.CreateOverride(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected first create 201, got %d", rr.Code)
	}

	body2, _ := json.Marshal(payload)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID.String()+"/budgets/overrides", bytes.NewBuffer(body2))
	req2 = req2.WithContext(auth.WithUser(req2.Context(), user))
	req2 = withURLParamAnalysis(req2, "projectId", projectID.String())

	rr2 := httptest.NewRecorder()
	handler.CreateOverride(rr2, req2)
	if rr2.Code != http.StatusConflict {
		t.Fatalf("expected conflict 409, got %d", rr2.Code)
	}
}
