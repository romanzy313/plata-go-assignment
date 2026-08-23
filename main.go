package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
)

func main() {
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		slog.Error("Error loading .env file", "error", err)
		os.Exit(1)
	}

	e := echo.New()
	e.Logger = slog.Default()

	e.GET("/health", func(c *echo.Context) error {
		return c.String(200, "OK")
	})

	e.Start(fmt.Sprintf(":%s", os.Getenv("PORT")))
}
