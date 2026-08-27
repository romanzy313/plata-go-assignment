package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

var errorResponseInternalError = errorResponse{Error: "Internal error"}

type server struct {
	e *echo.Echo
}

type handler struct {
	db serverDatabase
}

type errorResponse struct {
	Error string `json:"error"`
}

type updateResponse struct {
	UpdateId string `json:"updateId"`
}

type quoteResponse struct {
	UpdateId  string   `json:"updateId"`
	Pair      string   `json:"pair"`
	Status    string   `json:"status"`
	Price     *float64 `json:"price"`
	UpdatedAt *string  `json:"updatedAt"`
}

func newServer(logger *slog.Logger, db serverDatabase) *server {
	h := &handler{
		db: db,
	}

	e := echo.New()
	e.Logger = logger.With("source", "server")

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.GET("/health", func(c *echo.Context) error {
		return c.String(200, "OK")
	})
	e.POST("/update", h.update)
	e.GET("/quote/latest", h.getLatest)
	e.GET("/quote/:updateId", h.getByUpdateId)

	return &server{
		e: e,
	}
}

func (s *server) Run(ctx context.Context, port int) error {
	sc := echo.StartConfig{
		Address:         fmt.Sprintf(":%d", port),
		GracefulTimeout: 10 * time.Second,
		HideBanner:      true,
	}
	return sc.Start(ctx, s.e)
}

func (h *handler) update(c *echo.Context) error {
	idempotencyKey := c.Request().Header.Get("Idempotency-Key")
	parsedIdempotencyKey, err := uuid.Parse(idempotencyKey)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{
			Error: "Invalid Idempotency-Key header",
		})
	}

	pair := c.QueryParam("pair")
	base, quote, err := parseCurrencyPair(pair)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{
			Error: err.Error(),
		})
	}

	updateId, err := h.db.upsertUpdate(c.Request().Context(), &upsertUpdate{
		IdempotencyKey: parsedIdempotencyKey.String(),
		Base:           base,
		Quote:          quote,
	})
	if err != nil {
		c.Logger().Error("failed to upsert update", "error", err)
		return c.JSON(http.StatusInternalServerError, errorResponseInternalError)
	}

	return c.JSON(http.StatusOK, updateResponse{
		UpdateId: updateId,
	})
}

func (h *handler) getLatest(c *echo.Context) error {
	pair := c.QueryParam("pair")
	base, quote, err := parseCurrencyPair(pair)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{
			Error: err.Error(),
		})
	}

	update, err := h.db.getUpdateLatest(c.Request().Context(),
		base, quote)
	if err != nil {
		c.Logger().Error("failed to get latest", "error", err)
		return c.JSON(http.StatusInternalServerError, errorResponseInternalError)
	}
	if update == nil {
		return c.JSON(http.StatusNotFound, errorResponse{
			Error: "Update not found",
		})
	}

	return c.JSON(http.StatusOK, updateToQuoteResult(update))
}

func (h *handler) getByUpdateId(c *echo.Context) error {
	updateId := c.Param("updateId")
	if _, err := uuid.Parse(updateId); err != nil {
		return c.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid update ID"})
	}

	update, err := h.db.getUpdateById(c.Request().Context(), updateId)
	if err != nil {
		c.Logger().Error("failed to get by id", "error", err)
		return c.JSON(http.StatusInternalServerError, errorResponseInternalError)
	}
	if update == nil {
		return c.JSON(http.StatusNotFound, errorResponse{
			Error: "Update not found",
		})
	}
	return c.JSON(http.StatusOK, updateToQuoteResult(update))
}

func updateToQuoteResult(u *update) quoteResponse {
	// hide "processing" state from the client
	status := u.Status
	if status == updateStatusProcessing {
		status = updateStatusPending
	}

	result := quoteResponse{
		UpdateId: u.Id,
		Pair:     currencyPairString(u.Base, u.Quote),
		Status:   status.String(),
		Price:    u.Price,
	}

	if u.UpdatedAt != nil {
		updatedAt := u.UpdatedAt.UTC().Format(time.RFC3339)
		result.UpdatedAt = &updatedAt
	}
	return result
}
