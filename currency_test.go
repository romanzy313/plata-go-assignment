package main

import "testing"

func TestParseCurrencyPair(t *testing.T) {
	tests := []struct {
		pair      string
		wantBase  string
		wantQuote string
		wantErr   bool
	}{
		{pair: "EUR/MXN", wantBase: "EUR", wantQuote: "MXN"},
		{pair: "EUR/EUR", wantErr: true},
		{pair: "", wantErr: true},
		{pair: "too long of a value", wantErr: true},
		{pair: "A/B/C/D", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.pair, func(t *testing.T) {
			gotBase, gotQuote, err := parseCurrencyPair(test.pair)
			if test.wantErr && err == nil {
				t.Fatalf("wantErr but got no err")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if test.wantBase != gotBase {
				t.Fatalf("base: want %s; got %s", test.wantBase, gotBase)
			}
			if test.wantQuote != gotQuote {
				t.Fatalf("quote: want %s; got %s", test.wantQuote, gotQuote)
			}
		})
	}
}
