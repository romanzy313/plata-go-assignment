package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type handler struct {
	exchangeRateClient exchangeRateClient
	apiDatabase        apiDatabase
}

type errorResponse struct {
	Error string `json:"error"`
}

type rateRefreshResponse struct {
	UpdateId string `json:"updateId"`
}

type rateResponse struct {
	UpdateId  string   `json:"updateId"`
	Status    string   `json:"status"`
	Price     *float64 `json:"price"`
	UpdatedAt string   `json:"updatedAt"`
}

func newHTTPServer(h *handler) *echo.Echo {
	e := echo.New()
	e.Logger = slog.Default()

	e.Use(middleware.Recover())
	e.Use(middleware.RequestLogger())

	e.GET("/health", func(c *echo.Context) error {
		return c.String(200, "OK")
	})
	e.POST("/update", h.update)
	e.GET("/quote/latest", h.getLatest)
	e.GET("/quote/:updateId", h.getByUpdateId)

	// ugly and quick integration tests
	e.GET("/test/api", h.testApi)

	return e
}

// TODO: idempotency
func (h *handler) update(c *echo.Context) error {
	var request struct {
		Pair string `json:"pair"`
	}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{
			Error: "Invalid request",
		})
	}
	base, quote, err := parseCurrencyPair(request.Pair)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{
			Error: err.Error(),
		})
	}

	updateId := uuid.NewString()
	err = h.apiDatabase.createUpdate(c.Request().Context(), &newUpdate{
		Id:    updateId,
		Base:  base,
		Quote: quote,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "internal error",
		})
	}

	return c.JSON(http.StatusOK, rateRefreshResponse{
		UpdateId: updateId,
	})
}

func (h *handler) getLatest(c *echo.Context) error {
	var request struct {
		Pair string `query:"pair"`
	}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{
			Error: "Invalid request",
		})
	}
	base, quote, err := parseCurrencyPair(request.Pair)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{
			Error: err.Error(),
		})
	}
	update, err := h.apiDatabase.getUpdateLatest(c.Request().Context(),
		base, quote)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "internal error",
		})
	}

	return c.JSON(http.StatusOK, rateResponse{
		UpdateId:  update.Id,
		Status:    update.Status.String(),
		Price:     update.Price,
		UpdatedAt: update.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

func (h *handler) getByUpdateId(c *echo.Context) error {
	updateId := c.Param("updateId")
	if _, err := uuid.Parse(updateId); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid update ID"})
	}

	update, err := h.apiDatabase.getUpdateById(c.Request().Context(), updateId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errorResponse{
			Error: "internal error",
		})
	}
	if update == nil {
		return c.JSON(http.StatusNotFound, errorResponse{
			Error: "Update not found",
		})
	}

	return c.JSON(http.StatusOK, rateResponse{
		UpdateId:  update.Id,
		Status:    update.Status.String(),
		Price:     update.Price,
		UpdatedAt: update.UpdatedAt.UTC().Format(time.RFC3339),
	})
}

func (h *handler) testApi(c *echo.Context) error {
	snapshot, err := h.exchangeRateClient.latestSnapshot(
		c.Request().Context(), "EUR", []string{"USD", "MXN"})
	slog.Info("api test", "snapshot", snapshot, "err", err)
	if err != nil {
		return c.String(http.StatusInternalServerError, "not ok")
	}
	return c.String(http.StatusOK, "ok")
}
