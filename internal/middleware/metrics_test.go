package middleware

import (
	"testing"
	"time"
)

func TestPrometheusMetricsCollector(t *testing.T) {
	collector := NewPrometheusMetricsCollector()

	// Record some requests
	collector.RecordHTTPRequest("GET", "/api/events", 200, 100*time.Millisecond)
	collector.RecordHTTPRequest("GET", "/api/events", 200, 150*time.Millisecond)
	collector.RecordHTTPRequest("POST", "/api/tickets/purchase", 200, 200*time.Millisecond)
	collector.RecordHTTPRequest("GET", "/api/events", 500, 50*time.Millisecond)

	counts, durations := collector.GetMetrics()

	// Check counts
	getEventsKey := "GET:/api/events:200"
	if counts[getEventsKey] != 2 {
		t.Fatalf("Expected 2 GET /api/events requests, got %d", counts[getEventsKey])
	}

	postPurchaseKey := "POST:/api/tickets/purchase:200"
	if counts[postPurchaseKey] != 1 {
		t.Fatalf("Expected 1 POST /api/tickets/purchase request, got %d", counts[postPurchaseKey])
	}

	errorKey := "GET:/api/events:500"
	if counts[errorKey] != 1 {
		t.Fatalf("Expected 1 error request, got %d", counts[errorKey])
	}

	// Check durations
	expectedDuration := 250 * time.Millisecond
	if durations[getEventsKey] != expectedDuration {
		t.Fatalf("Expected duration %v, got %v", expectedDuration, durations[getEventsKey])
	}
}
