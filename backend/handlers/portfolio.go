package handlers

import (
	"github.com/gofiber/fiber/v2"

	"portfolio-index/middleware"
	"portfolio-index/models"
)

// ─── Watchlist ────────────────────────────────────────────────

// GET /user/watchlist
func (h *Handler) GetWatchlist(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	items, err := h.db.GetWatchlist(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch watchlist"})
	}
	return c.JSON(items)
}

// POST /user/watchlist/:symbol
func (h *Handler) AddWatchlist(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(400).JSON(fiber.Map{"error": "symbol required"})
	}
	if err := h.db.AddToWatchlist(c.Context(), userID, symbol); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to add to watchlist"})
	}
	return c.Status(201).JSON(fiber.Map{"ok": true})
}

// DELETE /user/watchlist/:symbol
func (h *Handler) RemoveWatchlist(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	symbol := c.Params("symbol")
	if err := h.db.RemoveFromWatchlist(c.Context(), userID, symbol); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to remove from watchlist"})
	}
	return c.JSON(fiber.Map{"ok": true})
}

// ─── Notes ────────────────────────────────────────────────────

// GET /user/notes/:symbol
func (h *Handler) GetNotes(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	symbol := c.Params("symbol")
	notes, err := h.db.GetNotes(c.Context(), userID, symbol)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch notes"})
	}
	return c.JSON(notes)
}

// POST /user/notes/:symbol
func (h *Handler) CreateNote(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	symbol := c.Params("symbol")
	var req models.UpsertNoteRequest
	if err := c.BodyParser(&req); err != nil || req.Content == "" {
		return c.Status(400).JSON(fiber.Map{"error": "content is required"})
	}
	note, err := h.db.CreateNote(c.Context(), userID, symbol, req.Content)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create note"})
	}
	return c.Status(201).JSON(note)
}

// PUT /user/notes/:id
func (h *Handler) UpdateNote(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	noteID := c.Params("id")
	var req models.UpsertNoteRequest
	if err := c.BodyParser(&req); err != nil || req.Content == "" {
		return c.Status(400).JSON(fiber.Map{"error": "content is required"})
	}
	note, err := h.db.UpdateNote(c.Context(), noteID, userID, req.Content)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update note"})
	}
	if note == nil {
		return c.Status(404).JSON(fiber.Map{"error": "note not found"})
	}
	return c.JSON(note)
}

// DELETE /user/notes/:id
func (h *Handler) DeleteNote(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	noteID := c.Params("id")
	if err := h.db.DeleteNote(c.Context(), noteID, userID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete note"})
	}
	return c.JSON(fiber.Map{"ok": true})
}

// ─── Portfolios ───────────────────────────────────────────────

// GET /user/portfolios
func (h *Handler) GetPortfolios(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	portfolios, err := h.db.GetPortfolios(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch portfolios"})
	}

	// For each portfolio, load holdings and calculate P&L
	for i := range portfolios {
		holdings, err := h.db.GetHoldings(c.Context(), portfolios[i].ID)
		if err != nil {
			continue
		}
		// Enrich with current prices
		for j := range holdings {
			idx, _ := h.db.GetIndex(c.Context(), holdings[j].Symbol)
			if idx != nil && idx.Price > 0 {
				holdings[j].CurrentPrice = idx.Price
				holdings[j].CurrentValue = idx.Price * holdings[j].Quantity
				holdings[j].PnL = holdings[j].CurrentValue - holdings[j].Cost
				if holdings[j].Cost > 0 {
					holdings[j].PnLPercent = holdings[j].PnL / holdings[j].Cost * 100
				}
				portfolios[i].TotalValue += holdings[j].CurrentValue
				portfolios[i].TotalCost += holdings[j].Cost
			} else {
				portfolios[i].TotalCost += holdings[j].Cost
			}
		}
		portfolios[i].Holdings = holdings
		portfolios[i].PnL = portfolios[i].TotalValue - portfolios[i].TotalCost
		if portfolios[i].TotalCost > 0 {
			portfolios[i].PnLPercent = portfolios[i].PnL / portfolios[i].TotalCost * 100
		}
	}
	return c.JSON(portfolios)
}

// POST /user/portfolios
func (h *Handler) CreatePortfolio(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	var req models.CreatePortfolioRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}
	p, err := h.db.CreatePortfolio(c.Context(), userID, req.Name, req.Description, req.Currency)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create portfolio"})
	}
	p.Holdings = []models.Holding{}
	return c.Status(201).JSON(p)
}

// DELETE /user/portfolios/:id
func (h *Handler) DeletePortfolio(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	pID := c.Params("id")
	if err := h.db.DeletePortfolio(c.Context(), pID, userID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete portfolio"})
	}
	return c.JSON(fiber.Map{"ok": true})
}

// ─── Holdings ─────────────────────────────────────────────────

// POST /user/portfolios/:id/holdings
func (h *Handler) AddHolding(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	pID := c.Params("id")

	ok, err := h.db.IsPortfolioOwner(c.Context(), pID, userID)
	if err != nil || !ok {
		return c.Status(403).JSON(fiber.Map{"error": "portfolio not found"})
	}

	var req models.CreateHoldingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Symbol == "" || req.Quantity <= 0 || req.AvgPrice < 0 {
		return c.Status(400).JSON(fiber.Map{"error": "symbol, quantity (>0) and avg_price (>=0) are required"})
	}

	holding, err := h.db.CreateHolding(c.Context(), pID, &req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to add holding"})
	}
	return c.Status(201).JSON(holding)
}

// DELETE /user/portfolios/:id/holdings/:hid
func (h *Handler) RemoveHolding(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	pID := c.Params("id")
	hID := c.Params("hid")

	ok, err := h.db.IsPortfolioOwner(c.Context(), pID, userID)
	if err != nil || !ok {
		return c.Status(403).JSON(fiber.Map{"error": "portfolio not found"})
	}

	if err := h.db.DeleteHolding(c.Context(), hID, pID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to remove holding"})
	}
	return c.JSON(fiber.Map{"ok": true})
}
