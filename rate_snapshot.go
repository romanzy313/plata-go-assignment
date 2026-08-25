package main

import (
	"fmt"
	"math"
	"time"
)

type rateSnapshot struct {
	Base      string
	Rates     map[string]float64
	UpdatedAt time.Time
}

func calculateExchangeRate(snapshot *rateSnapshot, base, quote string) (float64, error) {
	getRate := func(currency string) (float64, error) {
		if currency == snapshot.Base {
			return 1, nil
		}

		rate, ok := snapshot.Rates[currency]
		if !ok || rate <= 0 || math.IsNaN(rate) || math.IsInf(rate, 0) {
			return 0, fmt.Errorf(
				"snapshot is missing %s/%s rate",
				snapshot.Base,
				currency,
			)
		}

		return rate, nil
	}

	baseRate, err := getRate(base)
	if err != nil {
		return 0, err
	}

	quoteRate, err := getRate(quote)
	if err != nil {
		return 0, err
	}

	return quoteRate / baseRate, nil
}
