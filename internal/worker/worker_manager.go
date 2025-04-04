package worker

import (
	"sync"

	"github.com/guillermoballester/propagatorGo/internal/config"
)

// WorkManager manages the distribution of stocks to workers
type WorkManager struct {
	stocks     []config.Stock // List of stocks to process
	currentIdx int            // Current position in the stocks list
	mu         sync.Mutex     // To make operations thread-safe
}

// NewWorkManager creates a new work manager from a stock list
func NewWorkManager(stockList []config.Stock) *WorkManager {
	enabledStocks := make([]config.Stock, 0)
	for _, stock := range stockList {
		if stock.Enabled {
			enabledStocks = append(enabledStocks, stock)
		}
	}

	return &WorkManager{
		stocks:     enabledStocks,
		currentIdx: 0,
	}
}

// GetNextStock returns the next stock to process, cycling through the list
func (wm *WorkManager) GetNextStock() *config.Stock {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if len(wm.stocks) == 0 {
		return nil
	}

	stock := wm.stocks[wm.currentIdx]
	wm.currentIdx = (wm.currentIdx + 1) % len(wm.stocks)

	return &stock
}

// GetAllStocks returns all enabled stocks
func (wm *WorkManager) GetAllStocks() []config.Stock {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Return a copy to avoid potential data races
	result := make([]config.Stock, len(wm.stocks))
	copy(result, wm.stocks)

	return result
}

// GetStockCount returns the number of stocks managed
func (wm *WorkManager) GetStockCount() int {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	return len(wm.stocks)
}

// Reset resets the current index to the beginning
func (wm *WorkManager) Reset() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.currentIdx = 0
}

// NewWorkManagerFromConfig creates a WorkManager directly from the config
func NewWorkManagerFromConfig(cfg *config.Config) *WorkManager {
	return NewWorkManager(cfg.StockList.Stocks)
}
