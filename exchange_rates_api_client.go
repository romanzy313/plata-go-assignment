package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const exchangeratesapiBaseCurrency = "EUR"

type exchangeratesapiClient struct {
	apiKey string
	client *resty.Client
}
type exchangeratesapiErrorReponse struct {
	Success bool `json:"success"`
	Error   *struct {
		Code    any    `json:"code"`
		Message string `json:"message"`
		Info    string `json:"info"`
	} `json:"error"`
}

type exchangeratesapiLatestResponse struct {
	Success   bool               `json:"success"`
	Timestamp int64              `json:"timestamp"`
	Base      string             `json:"base"`
	Date      string             `json:"date"`
	Rates     map[string]float64 `json:"rates"`
}

func newExchangeratesapiClient(apiKey string) *exchangeratesapiClient {
	client := resty.New().
		SetBaseURL("https://api.exchangeratesapi.io").
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second).
		SetTimeout(10 * time.Second)

	return &exchangeratesapiClient{
		apiKey: apiKey,
		client: client,
	}
}

func (c *exchangeratesapiClient) latestSnapshot(ctx context.Context, symbols []string) (*rateSnapshot, error) {
	res, err := c.client.R().
		SetQueryParams(map[string]string{
			"access_key": c.apiKey,
			"base":       exchangeratesapiBaseCurrency,
			"symbols":    strings.Join(symbols, ","),
		}).
		SetHeader("Accept", "application/json").
		SetResult(&exchangeratesapiLatestResponse{}).
		SetError(&exchangeratesapiErrorReponse{}).
		Get("/v1/latest")
	if err != nil {
		return nil, err
	}
	if res.IsError() {
		body := res.Error().(*exchangeratesapiErrorReponse)
		return nil, fmt.Errorf("could not fetch latest snapshot: %d, %s",
			body.Error.Code, body.Error.Message)
	}

	body := res.Result().(*exchangeratesapiLatestResponse)

	return &rateSnapshot{
		Base:      exchangeratesapiBaseCurrency,
		Rates:     body.Rates,
		UpdatedAt: time.Unix(body.Timestamp, 0),
	}, nil
}
