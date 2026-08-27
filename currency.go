package main

import (
	"errors"
	"fmt"
	"strings"
)

var errCurrencyInvalid = errors.New("Invalid currency")
var errCurrencyNotSupported = errors.New("Currency is not supported")
var errCurrencyPairInvalid = errors.New("Invalid currency pair")

var supportedCurrenciesMap = map[string]struct{}{
	"USD": {},
	"EUR": {},
	"MXN": {},
}
var supportedCurrencies = supportedCurrenciesSlice()

func supportedCurrenciesSlice() []string {
	arr := []string{}
	for k := range supportedCurrenciesMap {
		arr = append(arr, k)
	}
	return arr
}

// parses, validates, and returns base and quote from currency pair
func parseCurrencyPair(pair string) (string, string, error) {
	if len(pair) != 7 {
		return "", "", errCurrencyPairInvalid
	}

	first, second, found := strings.Cut(pair, "/")
	if !found {
		return "", "", errCurrencyPairInvalid
	}

	validBase, err := validateCurrency(first)
	if err != nil {
		return "", "", err
	}
	validQuote, err := validateCurrency(second)
	if err != nil {
		return "", "", err
	}
	if validBase == validQuote {
		return "", "", errCurrencyPairInvalid
	}

	return validBase, validQuote, nil
}

func currencyPairString(base, quote string) string {
	return fmt.Sprintf("%s/%s", base, quote)
}

// validates and returns properly formatted currency code
func validateCurrency(value string) (string, error) {
	if len(value) != 3 {
		return "", errCurrencyInvalid
	}

	upper := strings.ToUpper(value)
	if _, ok := supportedCurrenciesMap[upper]; !ok {
		return "", errCurrencyNotSupported
	}

	return upper, nil
}
