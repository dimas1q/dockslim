package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dimas1q/dockslim/backend/internal/apitokens"
	"github.com/dimas1q/dockslim/backend/internal/auth"
	"github.com/dimas1q/dockslim/backend/internal/dashboard"
	"github.com/dimas1q/dockslim/backend/internal/featureflags"
	"github.com/google/uuid"
)

type apiTokenServiceStub struct {
	tokens []apitokens.Token
	err    error
}

type subscriptionServiceStub struct {
	featuresByUser map[uuid.UUID]featureflags.UserFeatures
	updateFn       func(ctx context.Context, input featureflags.UpdateSubscriptionInput) (featureflags.UserFeatures, error)
}

type dashboardServiceStub struct {
	values map[uuid.UUID]dashboard.AccountDashboard
}

func (s *apiTokenServiceStub) CreateToken(ctx context.Context, userID uuid.UUID, name string, expiresAt *time.Time) (apitokens.Token, string, error) {
	if s.err != nil {
		return apitokens.Token{}, "", s.err
	}
	token := apitokens.Token{ID: uuid.New(), UserID: userID, Name: name, CreatedAt: time.Now()}
	s.tokens = append(s.tokens, token)
	return token, "ds_api_token", nil
}

func (s *apiTokenServiceStub) ListTokens(ctx context.Context, userID uuid.UUID) ([]apitokens.Token, error) {
	return s.tokens, s.err
}

func (s *apiTokenServiceStub) RevokeToken(ctx context.Context, userID, tokenID uuid.UUID) error {
	return s.err
}

func (s *subscriptionServiceStub) GetUserFeatures(ctx context.Context, userID uuid.UUID) (featureflags.UserFeatures, error) {
	if s.featuresByUser == nil {
		return featureflags.UserFeatures{}, nil
	}
	if features, ok := s.featuresByUser[userID]; ok {
		return features, nil
	}
	return featureflags.UserFeatures{}, nil
}

func (s *subscriptionServiceStub) UpdateUserSubscription(ctx context.Context, input featureflags.UpdateSubscriptionInput) (featureflags.UserFeatures, error) {
	if s.updateFn != nil {
		return s.updateFn(ctx, input)
	}
	return featureflags.UserFeatures{}, nil
}

func (s *dashboardServiceStub) GetDashboard(ctx context.Context, userID uuid.UUID) (dashboard.AccountDashboard, error) {
	if s.values == nil {
		return dashboard.AccountDashboard{}, nil
	}
	if data, ok := s.values[userID]; ok {
		return data, nil
	}
	return dashboard.AccountDashboard{}, nil
}

func TestCreateAPITokenConflictReturns409(t *testing.T) {
	user := auth.User{ID: uuid.New()}
	service := &apiTokenServiceStub{err: apitokens.ErrNameConflict}
	handler := NewAccountHandler(nil, service)

	body, _ := json.Marshal(map[string]string{"name": "personal"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/account/api-tokens", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.CreateAPIToken(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestRevokeAPITokenNotFound(t *testing.T) {
	user := auth.User{ID: uuid.New()}
	service := &apiTokenServiceStub{err: apitokens.ErrTokenNotFound}
	handler := NewAccountHandler(nil, service)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/account/api-tokens/token/revoke", nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	tokenID := uuid.New().String()
	req = withURLParam(req, "tokenId", tokenID)
	rec := httptest.NewRecorder()

	handler.RevokeAPIToken(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUpdateProfileInvalidEmail(t *testing.T) {
	user := auth.User{ID: uuid.New(), Login: "demo", Email: "demo@example.com"}
	userStore := newMemoryUserStore()
	userStore.usersByID[user.ID.String()] = user
	userStore.usersByEmail[user.Email] = user
	userStore.usersByLogin[user.Login] = user
	tokenStore := newMemoryKeyStore()
	tokenManager := newTokenManager(t, tokenStore)
	authService := auth.NewService(userStore, tokenManager)
	handler := NewAccountHandler(authService, &apiTokenServiceStub{})

	body, _ := json.Marshal(map[string]string{"email": "bad-email"})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/account/me", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.UpdateProfile(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetSubscriptionReturnsPlan(t *testing.T) {
	user := auth.User{ID: uuid.New()}
	subscriptionSvc := &subscriptionServiceStub{
		featuresByUser: map[uuid.UUID]featureflags.UserFeatures{
			user.ID: {
				UserID:   user.ID,
				PlanID:   featureflags.PlanFree,
				PlanName: "Free",
				Status:   featureflags.SubscriptionStatusActive,
				Features: map[string]any{
					featureflags.FeatureAdvancedInsights: false,
					featureflags.FeatureHistoryDaysLimit: float64(30),
				},
			},
		},
	}
	handler := NewAccountHandler(nil, &apiTokenServiceStub{}, AccountHandlerOptions{
		SubscriptionService: subscriptionSvc,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/account/subscription", nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.GetSubscription(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	plan, ok := payload["plan"].(map[string]any)
	if !ok {
		t.Fatalf("expected plan object in response")
	}
	if plan["id"] != featureflags.PlanFree {
		t.Fatalf("expected free plan id, got %v", plan["id"])
	}
}

func TestUpdateSubscriptionRequiresAdminAndInternalToken(t *testing.T) {
	user := auth.User{ID: uuid.New(), IsAdmin: false}
	subscriptionSvc := &subscriptionServiceStub{
		featuresByUser: map[uuid.UUID]featureflags.UserFeatures{
			user.ID: {
				UserID:   user.ID,
				PlanID:   featureflags.PlanFree,
				PlanName: "Free",
				Status:   featureflags.SubscriptionStatusActive,
				Features: map[string]any{},
			},
		},
	}
	handler := NewAccountHandler(nil, &apiTokenServiceStub{}, AccountHandlerOptions{
		SubscriptionService:       subscriptionSvc,
		InternalSubscriptionToken: "internal-secret",
	})

	body, _ := json.Marshal(map[string]any{
		"user_id": user.ID.String(),
		"plan_id": featureflags.PlanPro,
		"status":  featureflags.SubscriptionStatusActive,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/internal/subscriptions", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.UpdateSubscription(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestUpdateSubscriptionAsAdmin(t *testing.T) {
	admin := auth.User{ID: uuid.New(), IsAdmin: true}
	targetUserID := uuid.New()
	subscriptionSvc := &subscriptionServiceStub{
		featuresByUser: map[uuid.UUID]featureflags.UserFeatures{
			admin.ID: {
				UserID:   admin.ID,
				PlanID:   featureflags.PlanTeam,
				PlanName: "Team",
				Status:   featureflags.SubscriptionStatusActive,
				Features: map[string]any{},
			},
		},
		updateFn: func(ctx context.Context, input featureflags.UpdateSubscriptionInput) (featureflags.UserFeatures, error) {
			return featureflags.UserFeatures{
				UserID:   input.UserID,
				PlanID:   input.PlanID,
				PlanName: "Pro",
				Status:   input.Status,
				Features: map[string]any{
					featureflags.FeatureExportJSON: true,
				},
			}, nil
		},
	}
	handler := NewAccountHandler(nil, &apiTokenServiceStub{}, AccountHandlerOptions{
		SubscriptionService:       subscriptionSvc,
		InternalSubscriptionToken: "internal-secret",
	})

	body, _ := json.Marshal(map[string]any{
		"user_id": targetUserID.String(),
		"plan_id": featureflags.PlanPro,
		"status":  featureflags.SubscriptionStatusActive,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/internal/subscriptions", bytes.NewBuffer(body))
	req = req.WithContext(auth.WithUser(req.Context(), admin))
	req.Header.Set("X-DockSlim-Internal-Token", "internal-secret")
	rec := httptest.NewRecorder()

	handler.UpdateSubscription(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGetDashboardUnauthorized(t *testing.T) {
	handler := NewAccountHandler(nil, &apiTokenServiceStub{}, AccountHandlerOptions{
		DashboardService: &dashboardServiceStub{},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/account/dashboard", nil)
	rec := httptest.NewRecorder()

	handler.GetDashboard(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestGetDashboardReturnsData(t *testing.T) {
	user := auth.User{ID: uuid.New()}
	handler := NewAccountHandler(nil, &apiTokenServiceStub{}, AccountHandlerOptions{
		DashboardService: &dashboardServiceStub{
			values: map[uuid.UUID]dashboard.AccountDashboard{
				user.ID: {
					Summary: dashboard.Summary{
						ProjectsTotal: 3,
						AnalysesTotal: 12,
					},
					Activity: dashboard.Activity{
						Last35Days: []dashboard.ActivityPoint{{Date: "2026-02-12", Count: 2, Level: 3}},
						RecentEvents: []dashboard.Event{
							{
								Type:        "analysis_completed",
								OccurredAt:  time.Now().UTC(),
								ProjectID:   uuid.New(),
								ProjectName: "core",
							},
						},
					},
				},
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/account/dashboard", nil)
	req = req.WithContext(auth.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()

	handler.GetDashboard(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload dashboard.AccountDashboard
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Summary.ProjectsTotal != 3 {
		t.Fatalf("expected projects_total=3, got %d", payload.Summary.ProjectsTotal)
	}
	if len(payload.Activity.RecentEvents) != 1 {
		t.Fatalf("expected one recent event, got %d", len(payload.Activity.RecentEvents))
	}
}
