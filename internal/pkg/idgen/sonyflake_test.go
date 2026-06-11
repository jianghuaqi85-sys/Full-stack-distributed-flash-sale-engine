package idgen

import (
	"testing"
)

func TestInit(t *testing.T) {
	err := Init()
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	if !IsInitialized() {
		t.Fatal("IsInitialized() should return true after Init()")
	}
}

func TestNextID(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	id1, err := NextID()
	if err != nil {
		t.Fatalf("NextID() failed: %v", err)
	}

	id2, err := NextID()
	if err != nil {
		t.Fatalf("NextID() failed: %v", err)
	}

	if id1 == id2 {
		t.Fatal("NextID() should return unique IDs")
	}

	if id1 == 0 || id2 == 0 {
		t.Fatal("NextID() should return non-zero IDs")
	}
}

func TestMustNextID(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	id := MustNextID()
	if id == 0 {
		t.Fatal("MustNextID() should return non-zero ID")
	}
}

func TestOrderNo(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	orderNo1 := OrderNo()
	orderNo2 := OrderNo()

	if orderNo1 == orderNo2 {
		t.Fatal("OrderNo() should return unique order numbers")
	}

	if len(orderNo1) != 18 { // "TK" + 16 hex chars
		t.Fatalf("OrderNo() length should be 18, got %d", len(orderNo1))
	}

	if orderNo1[:2] != "TK" {
		t.Fatalf("OrderNo() should start with 'TK', got '%s'", orderNo1[:2])
	}
}

func TestHealthCheck(t *testing.T) {
	// Before init
	sf = nil
	if err := HealthCheck(); err == nil {
		t.Fatal("HealthCheck() should fail before Init()")
	}

	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	if err := HealthCheck(); err != nil {
		t.Fatalf("HealthCheck() failed after Init(): %v", err)
	}
}

func TestExtractPodOrdinal(t *testing.T) {
	tests := []struct {
		name     string
		podName  string
		expected int
	}{
		{"valid ordinal", "seckill-0", 0},
		{"valid ordinal 1", "seckill-1", 1},
		{"valid ordinal 10", "seckill-10", 10},
		{"no ordinal", "seckill", -1},
		{"invalid ordinal", "seckill-abc", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPodOrdinal(tt.podName)
			if result != tt.expected {
				t.Fatalf("extractPodOrdinal(%s) = %d, want %d", tt.podName, result, tt.expected)
			}
		})
	}
}

func BenchmarkNextID(b *testing.B) {
	if err := Init(); err != nil {
		b.Fatalf("Init() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NextID()
		if err != nil {
			b.Fatalf("NextID() failed: %v", err)
		}
	}
}

func BenchmarkOrderNo(b *testing.B) {
	if err := Init(); err != nil {
		b.Fatalf("Init() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		OrderNo()
	}
}
