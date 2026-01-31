package budgets

import "testing"

func ptr(v int64) *int64 { return &v }

func TestEvaluateBudgetHardLimitFail(t *testing.T) {
	budget := &ResolvedBudget{HardLimitBytes: ptr(100)}
	res := EvaluateBudget(50, 120, budget)
	if res.Status != StatusFail {
		t.Fatalf("expected fail, got %s", res.Status)
	}
	if len(res.Reasons) == 0 || res.Reasons[0] != "exceeds hard limit" {
		t.Fatalf("expected hard limit reason")
	}
}

func TestEvaluateBudgetWarnDelta(t *testing.T) {
	budget := &ResolvedBudget{WarnDeltaBytes: ptr(10)}
	res := EvaluateBudget(100, 120, budget)
	if res.Status != StatusWarn {
		t.Fatalf("expected warn, got %s", res.Status)
	}
}

func TestEvaluateBudgetFailDelta(t *testing.T) {
	budget := &ResolvedBudget{FailDeltaBytes: ptr(10)}
	res := EvaluateBudget(100, 120, budget)
	if res.Status != StatusFail {
		t.Fatalf("expected fail, got %s", res.Status)
	}
}

func TestEvaluateBudgetNoGrowthOk(t *testing.T) {
	budget := &ResolvedBudget{FailDeltaBytes: ptr(10)}
	res := EvaluateBudget(120, 110, budget)
	if res.Status != StatusOK {
		t.Fatalf("expected ok, got %s", res.Status)
	}
}

func TestEvaluateBudgetReasonOrderDeterministic(t *testing.T) {
	budget := &ResolvedBudget{FailDeltaBytes: ptr(5), HardLimitBytes: ptr(100)}
	res := EvaluateBudget(90, 110, budget)
	if res.Status != StatusFail {
		t.Fatalf("expected fail, got %s", res.Status)
	}
	if len(res.Reasons) != 2 {
		t.Fatalf("expected 2 reasons, got %d", len(res.Reasons))
	}
	// sorted lexicographically for stability
	if res.Reasons[0] != "exceeds hard limit" {
		t.Fatalf("expected hard limit first, got %s", res.Reasons[0])
	}
}
