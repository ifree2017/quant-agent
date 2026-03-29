package handler

import (
	"quant-agent/internal/model"
)

// StoreInterface defines the store methods used by handlers.
// Both *store.Store (production) and *mockStore (test) implement this interface.
type StoreInterface interface {
	ListStrategies(userID string) ([]model.Strategy, error)
	GetStrategy(id string) (*model.Strategy, error)
	DeleteStrategy(id string) error
	SaveStrategy(strategy *model.Strategy) error
	GetBacktest(id string) (*model.BacktestResult, error)
	SaveBacktest(result *model.BacktestResult) error
}

// globalStoreInterface is the interface-based global store (for test injection).
var globalStoreInterface StoreInterface

// backtestStoreInterface is the interface-based backtest store (for test injection).
var backtestStoreInterface StoreInterface

// SetStoreInterface injects a store via interface (for testing).
// The existing SetStore(s *store.Store) remains for production use.
func SetStoreInterface(s StoreInterface) {
	globalStoreInterface = s
}

// SetBacktestStoreInterface injects a backtest store via interface (for testing).
// The existing SetBacktestStore(s *store.Store) remains for production use.
func SetBacktestStoreInterface(s StoreInterface) {
	backtestStoreInterface = s
}

// storeGetter abstracts access to the store for handlers that support both
// the concrete *store.Store and the StoreInterface.
type storeGetter interface {
	GetStore() StoreInterface
}
