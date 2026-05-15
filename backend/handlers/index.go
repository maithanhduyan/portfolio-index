package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"encoding/json"
	"fmt"
	"net/http"
	"portfolio-index/config"
	"portfolio-index/models"
	"portfolio-index/repository"
	"portfolio-index/services"
)

// Handler holds dependencies for HTTP route handlers.
type Handler struct {
	db    *repository.TimescaleDB
	cache *services.RedisCache
	cfg   *config.Config
}

func New(db *repository.TimescaleDB, cache *services.RedisCache, cfg *config.Config) *Handler {
	return &Handler{db: db, cache: cache, cfg: cfg}
}

// GET /api/v1/indices
func (h *Handler) GetAllIndices(c *fiber.Ctx) error {
	ctx := c.Context()

	var indices []models.Index
	if err := h.cache.GetAllIndices(ctx, &indices); err == nil {
		return c.JSON(indices)
	}

	var err error
	indices, err = h.db.GetAllIndices(ctx)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch indices")
	}

	_ = h.cache.SetAllIndices(ctx, indices)
	return c.JSON(indices)
}

// GET /api/v1/indices/:symbol
func (h *Handler) GetIndex(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	ctx := c.Context()

	var idx models.Index
	if err := h.cache.GetIndex(ctx, symbol, &idx); err == nil {
		return c.JSON(idx)
	}

	result, err := h.db.GetIndex(ctx, symbol)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "index not found")
	}

	_ = h.cache.SetIndex(ctx, symbol, result)
	return c.JSON(result)
}

// GET /api/v1/indices/:symbol/candles?interval=1h&from=<unix>&to=<unix>&limit=300
func (h *Handler) GetCandles(c *fiber.Ctx) error {
	symbol := c.Params("symbol")

	var q models.CandleQuery
	if err := c.QueryParser(&q); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid query params")
	}

	if q.Interval == "" {
		q.Interval = "1h"
	}
	if q.Limit == 0 {
		q.Limit = 300
	}

	now := time.Now()
	to := now
	from := now.Add(-30 * 24 * time.Hour)

	if q.To > 0 {
		to = time.Unix(q.To, 0)
	}
	if q.From > 0 {
		from = time.Unix(q.From, 0)
	}

	candles, err := h.db.GetCandles(c.Context(), symbol, q.Interval, from, to, q.Limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch candles")
	}

	return c.JSON(candles)
}

// GET /api/v1/indices/:symbol/stats
func (h *Handler) GetStats(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	ctx := c.Context()

	var stats models.IndexStats
	if err := h.cache.GetStats(ctx, symbol, &stats); err == nil {
		return c.JSON(stats)
	}

	result, err := h.db.GetStats(ctx, symbol)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch stats")
	}

	_ = h.cache.SetStats(ctx, symbol, result)
	return c.JSON(result)
}

// GET /api/v1/indices/:symbol/ai-analysis
// Proxies to the Python AI service.
func (h *Handler) GetAIAnalysis(c *fiber.Ctx) error {
	symbol := c.Params("symbol")

	url := fmt.Sprintf("%s/analyze/%s", h.cfg.AIServiceUrl, symbol)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to build AI request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "AI service unavailable")
	}
	defer resp.Body.Close()

	var result any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to decode AI response")
	}

	return c.JSON(result)
}
