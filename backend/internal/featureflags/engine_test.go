package featureflags

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeStore struct {
	plans  map[string]Plan
	subs   map[uuid.UUID]subscriptionWithPlan
	admins map[uuid.UUID]bool
}

func (f *fakeStore) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	return f.admins[userID], nil
}

func (f *fakeStore) GetUserSubscriptionWithPlan(ctx context.Context, userID uuid.UUID) (subscriptionWithPlan, error) {
	sub, ok := f.subs[userID]
	if !ok {
		return subscriptionWithPlan{}, ErrSubscriptionNotFound
	}
	return sub, nil
}

func (f *fakeStore) GetPlanByID(ctx context.Context, planID string) (Plan, error) {
	plan, ok := f.plans[planID]
	if !ok {
		return Plan{}, ErrPlanNotFound
	}
	return plan, nil
}

func (f *fakeStore) ListPlans(ctx context.Context) ([]Plan, error) {
	out := make([]Plan, 0, len(f.plans))
	for _, plan := range f.plans {
		out = append(out, plan)
	}
	return out, nil
}

func (f *fakeStore) UpsertUserSubscription(ctx context.Context, input UpdateSubscriptionInput) (UserSubscription, error) {
	plan, ok := f.plans[input.PlanID]
	if !ok {
		return UserSubscription{}, ErrPlanNotFound
	}
	sub := subscriptionWithPlan{
		UserSubscription: UserSubscription{
			UserID:     input.UserID,
			PlanID:     input.PlanID,
			Status:     input.Status,
			ValidUntil: input.ValidUntil,
		},
		PlanName: plan.Name,
		Features: cloneFeatures(plan.Features),
	}
	f.subs[input.UserID] = sub
	return sub.UserSubscription, nil
}

func TestGetUserFeaturesDefaultsToFreeWhenSubscriptionMissing(t *testing.T) {
	store := newFakeStore()
	engine := NewEngine(store)
	userID := uuid.New()

	features, err := engine.GetUserFeatures(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if features.PlanID != PlanFree {
		t.Fatalf("expected free plan, got %q", features.PlanID)
	}
	if enabled, _ := features.FeatureValue(FeatureAdvancedInsights); FeatureEnabled(enabled) {
		t.Fatalf("expected advanced insights to be disabled for default free plan")
	}
}

func TestHasFeatureForProPlan(t *testing.T) {
	store := newFakeStore()
	engine := NewEngine(store)
	userID := uuid.New()
	store.subs[userID] = subscriptionWithPlan{
		UserSubscription: UserSubscription{
			UserID: userID,
			PlanID: PlanPro,
			Status: SubscriptionStatusActive,
		},
		PlanName: store.plans[PlanPro].Name,
		Features: cloneFeatures(store.plans[PlanPro].Features),
	}

	enabled, err := engine.HasFeature(context.Background(), userID, FeatureExportJSON)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !enabled {
		t.Fatalf("expected export_json to be enabled for pro plan")
	}
}

func TestExpiredSubscriptionFallsBackToFree(t *testing.T) {
	store := newFakeStore()
	engine := NewEngine(store)
	now := time.Date(2026, 2, 12, 10, 0, 0, 0, time.UTC)
	engine.nowFn = func() time.Time { return now }

	userID := uuid.New()
	expiredAt := now.Add(-time.Hour)
	store.subs[userID] = subscriptionWithPlan{
		UserSubscription: UserSubscription{
			UserID:     userID,
			PlanID:     PlanPro,
			Status:     SubscriptionStatusActive,
			ValidUntil: &expiredAt,
		},
		PlanName: store.plans[PlanPro].Name,
		Features: cloneFeatures(store.plans[PlanPro].Features),
	}

	features, err := engine.GetUserFeatures(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if features.PlanID != PlanFree {
		t.Fatalf("expected expired subscription to fallback to free, got %q", features.PlanID)
	}
}

func TestAdminGetsAllFeatures(t *testing.T) {
	store := newFakeStore()
	engine := NewEngine(store)

	adminID := uuid.New()
	store.admins[adminID] = true
	store.subs[adminID] = subscriptionWithPlan{
		UserSubscription: UserSubscription{
			UserID: adminID,
			PlanID: PlanFree,
			Status: SubscriptionStatusActive,
		},
		PlanName: store.plans[PlanFree].Name,
		Features: cloneFeatures(store.plans[PlanFree].Features),
	}

	features, err := engine.GetUserFeatures(context.Background(), adminID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if features.PlanID != PlanTeam {
		t.Fatalf("expected admin plan to resolve as %q, got %q", PlanTeam, features.PlanID)
	}

	checks := []string{
		FeatureAdvancedInsights,
		FeatureExportPDF,
		FeatureExportJSON,
		FeatureBaselineSLA,
		FeatureTeamManagement,
		FeatureSharedProjects,
		FeatureAdvancedTrends,
	}
	for _, feature := range checks {
		enabled, ok := features.FeatureValue(feature)
		if !ok || !FeatureEnabled(enabled) {
			t.Fatalf("expected feature %q to be enabled for admin", feature)
		}
	}
}

func TestUpdateUserSubscription(t *testing.T) {
	store := newFakeStore()
	engine := NewEngine(store)
	userID := uuid.New()
	validUntil := time.Now().Add(24 * time.Hour).UTC()

	result, err := engine.UpdateUserSubscription(context.Background(), UpdateSubscriptionInput{
		UserID:     userID,
		PlanID:     PlanTeam,
		Status:     SubscriptionStatusActive,
		ValidUntil: &validUntil,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.PlanID != PlanTeam {
		t.Fatalf("expected team plan, got %q", result.PlanID)
	}
	if enabled, _ := result.FeatureValue(FeatureSharedProjects); !FeatureEnabled(enabled) {
		t.Fatalf("expected shared_projects enabled for team plan")
	}
}

func newFakeStore() *fakeStore {
	plans := map[string]Plan{
		PlanFree: {
			ID:   PlanFree,
			Name: "Free",
			Features: map[string]any{
				FeatureBasicAnalysis:    true,
				FeatureHistoryDaysLimit: float64(30),
				FeatureAdvancedInsights: false,
				FeatureExportPDF:        false,
				FeatureExportJSON:       false,
				FeatureCIComments:       CICommentsModeLimited,
				FeatureBaselineSLA:      false,
				FeatureTeamManagement:   false,
				FeatureSharedProjects:   false,
				FeatureAdvancedTrends:   false,
			},
		},
		PlanPro: {
			ID:   PlanPro,
			Name: "Pro",
			Features: map[string]any{
				FeatureBasicAnalysis:    true,
				FeatureHistoryDaysLimit: nil,
				FeatureAdvancedInsights: true,
				FeatureExportPDF:        true,
				FeatureExportJSON:       true,
				FeatureCIComments:       true,
				FeatureBaselineSLA:      true,
				FeatureTeamManagement:   false,
				FeatureSharedProjects:   false,
				FeatureAdvancedTrends:   false,
			},
		},
		PlanTeam: {
			ID:   PlanTeam,
			Name: "Team",
			Features: map[string]any{
				FeatureBasicAnalysis:    true,
				FeatureHistoryDaysLimit: nil,
				FeatureAdvancedInsights: true,
				FeatureExportPDF:        true,
				FeatureExportJSON:       true,
				FeatureCIComments:       true,
				FeatureBaselineSLA:      true,
				FeatureTeamManagement:   true,
				FeatureSharedProjects:   true,
				FeatureAdvancedTrends:   true,
			},
		},
	}
	return &fakeStore{
		plans:  plans,
		subs:   map[uuid.UUID]subscriptionWithPlan{},
		admins: map[uuid.UUID]bool{},
	}
}
