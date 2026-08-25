package main

import (
	"testing"
)

func TestValidSnapshot(t *testing.T) {
	snapshot := rateSnapshot{
		Base: "EUR",
		Rates: map[string]float64{
			"USD": 2,
			"MXN": 4,
		},
	}

	tests := []struct {
		base     string
		quote    string
		wantRate float64
		wantErr  bool
	}{
		{base: "EUR", quote: "MXN", wantRate: 4},
		{base: "MXN", quote: "EUR", wantRate: 0.25},
		{base: "USD", quote: "MXN", wantRate: 2},
		{base: "USD", quote: "oups", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.base+"_"+test.quote, func(t *testing.T) {
			gotRate, err := calculateExchangeRate(&snapshot, test.base, test.quote)
			if test.wantErr && err == nil {
				t.Fatalf("wantErr but got no err")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if test.wantRate != gotRate {
				t.Fatalf("want %f; got %f", test.wantRate, gotRate)
			}
		})
	}
}
