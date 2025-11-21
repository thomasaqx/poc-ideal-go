package storage

import "sync"

type Watchlist struct {
	mu     sync.RWMutex
	assets map[string]bool
}

func NewWatchlist() *Watchlist {
	return &Watchlist{
		assets: make(map[string]bool),
	}
}

// Add inserts a symbol if it is not already present.
func (w *Watchlist) Add(symbol string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.assets[symbol] {
		return false
	}

	w.assets[symbol] = true
	return true
}

func (w *Watchlist) GetAll() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	symbols := make([]string, 0, len(w.assets))
	for symbol := range w.assets {
		symbols = append(symbols, symbol)
	}
	return symbols
}
