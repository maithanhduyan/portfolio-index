package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"portfolio-index/models"
)

// ─── Users ────────────────────────────────────────────────────

func (db *TimescaleDB) CreateUser(ctx context.Context, email, name, hash string) (*models.User, error) {
	var u models.User
	err := db.pool.QueryRow(ctx, `
		INSERT INTO users (email, name, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, created_at
	`, email, name, hash).Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (db *TimescaleDB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := db.pool.QueryRow(ctx, `
		SELECT id, email, name, password_hash, created_at
		FROM users WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ─── Watchlist ────────────────────────────────────────────────

func (db *TimescaleDB) GetWatchlist(ctx context.Context, userID string) ([]models.WatchlistItem, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, user_id, symbol, created_at
		FROM watchlist WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.WatchlistItem, 0)
	for rows.Next() {
		var w models.WatchlistItem
		if err := rows.Scan(&w.ID, &w.UserID, &w.Symbol, &w.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, rows.Err()
}

func (db *TimescaleDB) AddToWatchlist(ctx context.Context, userID, symbol string) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO watchlist (user_id, symbol)
		VALUES ($1, $2)
		ON CONFLICT (user_id, symbol) DO NOTHING
	`, userID, symbol)
	return err
}

func (db *TimescaleDB) RemoveFromWatchlist(ctx context.Context, userID, symbol string) error {
	_, err := db.pool.Exec(ctx, `
		DELETE FROM watchlist WHERE user_id = $1 AND symbol = $2
	`, userID, symbol)
	return err
}

// ─── Notes ────────────────────────────────────────────────────

func (db *TimescaleDB) GetNotes(ctx context.Context, userID, symbol string) ([]models.Note, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT id, user_id, symbol, content, created_at, updated_at
		FROM notes WHERE user_id = $1 AND symbol = $2 ORDER BY updated_at DESC
	`, userID, symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := make([]models.Note, 0)
	for rows.Next() {
		var n models.Note
		if err := rows.Scan(&n.ID, &n.UserID, &n.Symbol, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (db *TimescaleDB) CreateNote(ctx context.Context, userID, symbol, content string) (*models.Note, error) {
	var n models.Note
	err := db.pool.QueryRow(ctx, `
		INSERT INTO notes (user_id, symbol, content)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, symbol, content, created_at, updated_at
	`, userID, symbol, content).Scan(&n.ID, &n.UserID, &n.Symbol, &n.Content, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (db *TimescaleDB) UpdateNote(ctx context.Context, noteID, userID, content string) (*models.Note, error) {
	var n models.Note
	err := db.pool.QueryRow(ctx, `
		UPDATE notes SET content = $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3
		RETURNING id, user_id, symbol, content, created_at, updated_at
	`, content, noteID, userID).Scan(&n.ID, &n.UserID, &n.Symbol, &n.Content, &n.CreatedAt, &n.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &n, err
}

func (db *TimescaleDB) DeleteNote(ctx context.Context, noteID, userID string) error {
	_, err := db.pool.Exec(ctx, `
		DELETE FROM notes WHERE id = $1 AND user_id = $2
	`, noteID, userID)
	return err
}
