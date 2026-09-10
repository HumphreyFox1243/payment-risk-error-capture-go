package main

import "testing"

func TestAssessRiskDecision(t *testing.T) {
	tests := []struct {
		name  string
		event PaymentEvent
		want  Decision
	}{
		{"valid small payment", PaymentEvent{ID: "p1", CustomerID: "c1", AmountCents: 1200}, Approve},
		{"repeated attempts", PaymentEvent{ID: "p2", CustomerID: "c1", AmountCents: 1200, Attempts: 3}, Review},
		{"missing amount", PaymentEvent{ID: "p3", CustomerID: "c1"}, Decline},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := assess(tt.event); got != tt.want {
				t.Fatalf("assess() = %q, want %q", got, tt.want)
			}
		})
	}
}
