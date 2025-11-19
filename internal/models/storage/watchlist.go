package storage

import (
	"sync"
)

type WatchList struct {
	mu     sync.RWMutex
	assets map[string]bool
}

func NewWatchList() *WatchList {
	return &WatchList{
		assets: make(map[string]bool),
	}
}

func (w *WatchList) Add(symbol string) bool {
	w.mu.Lock() //block writer
	defer w.mu.Unlock()

	if w.assets[symbol] {
		return false
	}

	w.assets[symbol] = true
	return true
}

func (w *WatchList) GetAll() []string {
	w.mu.RLock()         
	defer w.mu.RUnlock() // unlock

	symbols := make([]string, 0, len(w.assets))
	for symbol := range w.assets {
		symbols = append(symbols, symbol)
	}
	return symbols
}
