package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseHistoryFilterDateRangeInclusive(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/123/history?from=2026-02-01&to=2026-02-01", nil)
	filter, err := parseHistoryFilter(req, 100)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if filter.From == nil || filter.To == nil {
		t.Fatalf("expected from/to to be set")
	}

	start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 1, 23, 59, 59, 999999999, time.UTC)

	if !filter.From.Equal(start) {
		t.Fatalf("expected from %v, got %v", start, filter.From)
	}
	if !filter.To.Equal(end) {
		t.Fatalf("expected to %v, got %v", end, filter.To)
	}
}

func TestParseHistoryFilterRFC3339(t *testing.T) {
	value := "2026-02-01T15:04:05Z"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/123/history?from="+value+"&to="+value, nil)
	filter, err := parseHistoryFilter(req, 100)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected, _ := time.Parse(time.RFC3339, value)
	if filter.From == nil || !filter.From.Equal(expected) {
		t.Fatalf("expected from %v, got %v", expected, filter.From)
	}
	if filter.To == nil || !filter.To.Equal(expected) {
		t.Fatalf("expected to %v, got %v", expected, filter.To)
	}
}
