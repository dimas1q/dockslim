package budgets

import "sort"

const (
	StatusOK   = "ok"
	StatusWarn = "warn"
	StatusFail = "fail"
)

// EvaluateBudget applies thresholds to from/to totals and returns verdict.
func EvaluateBudget(fromTotalBytes, toTotalBytes int64, budget *ResolvedBudget) EvaluationResult {
	delta := toTotalBytes - fromTotalBytes
	result := EvaluationResult{
		Status:       StatusOK,
		Reasons:      []string{},
		DeltaBytes:   delta,
		ToTotalBytes: toTotalBytes,
	}

	var warnThreshold, failThreshold, hardLimit *int64
	if budget != nil {
		warnThreshold = budget.WarnDeltaBytes
		failThreshold = budget.FailDeltaBytes
		hardLimit = budget.HardLimitBytes
	}
	result.WarnDeltaBytes = warnThreshold
	result.FailDeltaBytes = failThreshold
	result.HardLimitBytes = hardLimit

	reasons := make([]string, 0, 3)
	status := StatusOK

	if hardLimit != nil && toTotalBytes > *hardLimit {
		reasons = append(reasons, "exceeds hard limit")
		status = StatusFail
	}

	if delta > 0 {
		if failThreshold != nil && delta > *failThreshold {
			reasons = append(reasons, "size regression over fail threshold")
			status = StatusFail
		} else if warnThreshold != nil && delta > *warnThreshold {
			reasons = append(reasons, "size regression over warning threshold")
			if status != StatusFail {
				status = StatusWarn
			}
		}
	}

	sort.Strings(reasons)

	result.Status = status
	result.Reasons = reasons
	return result
}
