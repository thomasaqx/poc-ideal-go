package storage

import (
	"database/sql"
	"fmt"
)

// Watchlist stores a database connection instead of an in-memory map.
type Watchlist struct {
	db *sql.DB
}

// NewWatchlist wires an existing sql.DB instance into the repository.
func NewWatchlist(db *sql.DB) *Watchlist {
	return &Watchlist{db: db}
}

// Add persists a new symbol. Returns true when inserted, false if it already existed.
func (w *Watchlist) Add(symbol string) (bool, error) {
	const query = "INSERT IGNORE INTO watchlist (symbol) VALUES (?)"

	result, err := w.db.Exec(query, symbol)
	if err != nil {
		return false, fmt.Errorf("failed to insert watchlist symbol: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to inspect affected rows: %w", err)
	}

	return rowsAffected > 0, nil
}

// GetAll fetches every tracked symbol from the database ordered as returned by MySQL.
func (w *Watchlist) GetAll() ([]string, error) {
	const query = "SELECT symbol FROM watchlist"

	rows, err := w.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch watchlist: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var sym string
		if err := rows.Scan(&sym); err != nil {
			return nil, fmt.Errorf("failed to scan watchlist row: %w", err)
		}
		symbols = append(symbols, sym)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cursor error while fetching watchlist: %w", err)
	}

	return symbols, nil
}
