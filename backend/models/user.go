package models

import "time"

// User account
type User struct {
	ID           string    `json:"id"         db:"id"`
	Email        string    `json:"email"      db:"email"`
	Name         string    `json:"name"       db:"name"`
	PasswordHash string    `json:"-"          db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Portfolio holds a named collection of assets.
type Portfolio struct {
	ID          string    `json:"id"          db:"id"`
	UserID      string    `json:"user_id"     db:"user_id"`
	Name        string    `json:"name"        db:"name"`
	Description string    `json:"description" db:"description"`
	Currency    string    `json:"currency"    db:"currency"`
	CreatedAt   time.Time `json:"created_at"  db:"created_at"`

	// Calculated at query time
	TotalValue float64   `json:"total_value"`
	TotalCost  float64   `json:"total_cost"`
	PnL        float64   `json:"pnl"`
	PnLPercent float64   `json:"pnl_percent"`
	Holdings   []Holding `json:"holdings,omitempty"`
}

// Holding is one position inside a Portfolio.
type Holding struct {
	ID          string    `json:"id"           db:"id"`
	PortfolioID string    `json:"portfolio_id" db:"portfolio_id"`
	Symbol      string    `json:"symbol"       db:"symbol"`
	Quantity    float64   `json:"quantity"     db:"quantity"`
	AvgPrice    float64   `json:"avg_price"    db:"avg_price"`
	BuyDate     *string   `json:"buy_date"     db:"buy_date"`
	Note        string    `json:"note"         db:"note"`
	CreatedAt   time.Time `json:"created_at"   db:"created_at"`

	// Populated from current index price
	CurrentPrice float64 `json:"current_price"`
	CurrentValue float64 `json:"current_value"`
	Cost         float64 `json:"cost"`
	PnL          float64 `json:"pnl"`
	PnLPercent   float64 `json:"pnl_percent"`
}

// WatchlistItem is a symbol saved by a user.
type WatchlistItem struct {
	ID        string    `json:"id"         db:"id"`
	UserID    string    `json:"user_id"    db:"user_id"`
	Symbol    string    `json:"symbol"     db:"symbol"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Note is a personal annotation on a market symbol.
type Note struct {
	ID        string    `json:"id"         db:"id"`
	UserID    string    `json:"user_id"    db:"user_id"`
	Symbol    string    `json:"symbol"     db:"symbol"`
	Content   string    `json:"content"    db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ─── Request / Response DTOs ──────────────────────────────────

type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreatePortfolioRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Currency    string `json:"currency"`
}

type CreateHoldingRequest struct {
	Symbol   string  `json:"symbol"`
	Quantity float64 `json:"quantity"`
	AvgPrice float64 `json:"avg_price"`
	BuyDate  *string `json:"buy_date"`
	Note     string  `json:"note"`
}

type UpsertNoteRequest struct {
	Content string `json:"content"`
}
