package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"portfolio-index/models"
)

// ─── Portfolios ───────────────────────────────────────────────

func (db *TimescaleDB) GetPortfolios(ctx context.Context, userID string) ([]models.Portfolio, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, user_id, name, description, currency, created_at
		FROM portfolios WHERE user_id = $1 ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	portfolios := make([]models.Portfolio, 0)
	for rows.Next() {
		var p models.Portfolio
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Currency, &p.CreatedAt); err != nil {
			return nil, err
		}
		portfolios = append(portfolios, p)
	}
	return portfolios, rows.Err()
}

func (db *TimescaleDB) CreatePortfolio(ctx context.Context, userID, name, description, currency string) (*models.Portfolio, error) {
	var p models.Portfolio
	err := db.pool.QueryRow(ctx, `
		INSERT INTO portfolios (user_id, name, description, currency)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, name, description, currency, created_at
	`, userID, name, description, currency).
		Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Currency, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (db *TimescaleDB) DeletePortfolio(ctx context.Context, portfolioID, userID string) error {
	_, err := db.pool.Exec(ctx, `
		DELETE FROM portfolios WHERE id = $1 AND user_id = $2
	`, portfolioID, userID)
	return err
}

// ─── Holdings ─────────────────────────────────────────────────

func (db *TimescaleDB) GetHoldings(ctx context.Context, portfolioID string) ([]models.Holding, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, portfolio_id, symbol, quantity, avg_price,
		       to_char(buy_date, 'YYYY-MM-DD'), note, created_at
		FROM holdings WHERE portfolio_id = $1 ORDER BY created_at ASC
	`, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	holdings := make([]models.Holding, 0)
	for rows.Next() {
		var h models.Holding
		if err := rows.Scan(
			&h.ID, &h.PortfolioID, &h.Symbol, &h.Quantity, &h.AvgPrice,
			&h.BuyDate, &h.Note, &h.CreatedAt,
		); err != nil {
			return nil, err
		}
		h.Cost = h.Quantity * h.AvgPrice
		holdings = append(holdings, h)
	}
	return holdings, rows.Err()
}

func (db *TimescaleDB) CreateHolding(ctx context.Context, portfolioID string, req *models.CreateHoldingRequest) (*models.Holding, error) {
	var h models.Holding
	err := db.pool.QueryRow(ctx, `
		INSERT INTO holdings (portfolio_id, symbol, quantity, avg_price, buy_date, note)
		VALUES ($1, $2, $3, $4,
		        CASE WHEN $5::text = '' THEN NULL ELSE $5::date END,
		        $6)
		RETURNING id, portfolio_id, symbol, quantity, avg_price,
		          to_char(buy_date, 'YYYY-MM-DD'), note, created_at
	`, portfolioID, req.Symbol, req.Quantity, req.AvgPrice,
		nullableStr(req.BuyDate), req.Note).
		Scan(&h.ID, &h.PortfolioID, &h.Symbol, &h.Quantity, &h.AvgPrice,
			&h.BuyDate, &h.Note, &h.CreatedAt)
	if err != nil {
		return nil, err
	}
	h.Cost = h.Quantity * h.AvgPrice
	return &h, nil
}

func (db *TimescaleDB) DeleteHolding(ctx context.Context, holdingID, portfolioID string) error {
	_, err := db.pool.Exec(ctx, `
		DELETE FROM holdings WHERE id = $1 AND portfolio_id = $2
	`, holdingID, portfolioID)
	return err
}

// IsPortfolioOwner checks that a portfolio belongs to a user.
func (db *TimescaleDB) IsPortfolioOwner(ctx context.Context, portfolioID, userID string) (bool, error) {
	var exists bool
	err := db.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM portfolios WHERE id = $1 AND user_id = $2)
	`, portfolioID, userID).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return exists, err
}

func nullableStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
