package tracker

import (
	"context"
	"fmt"

	"github.com/makalin/pricetrek/internal/config"
	"github.com/makalin/pricetrek/internal/logger"
	"github.com/makalin/pricetrek/internal/providers"
	"github.com/makalin/pricetrek/internal/storage"
)

type Tracker struct {
	config  *config.Config
	storage storage.Storage
	logger  *logger.Logger
}

func New(cfg *config.Config, store storage.Storage, log *logger.Logger) *Tracker {
	return &Tracker{
		config:  cfg,
		storage: store,
		logger:  log,
	}
}

func (t *Tracker) TrackItem(ctx context.Context, item config.ItemConfig) error {
	t.logger.Debug("Tracking item", "id", item.ID, "name", item.Name)

	// Get provider
	provider, err := providers.GetProvider(item.Provider, t.config.Defaults)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	// Fetch price
	sample, err := provider.Fetch(ctx, item)
	if err != nil {
		return fmt.Errorf("failed to fetch price: %w", err)
	}

	// Save to storage
	if err := t.storage.SavePrice(ctx, item.ID, sample.Price, sample.Currency, sample.Meta); err != nil {
		return fmt.Errorf("failed to save price: %w", err)
	}

	t.logger.Info("Price tracked", 
		"item", item.ID, 
		"price", sample.Price, 
		"currency", sample.Currency,
	)

	return nil
}

func (t *Tracker) TrackAll(ctx context.Context) error {
	t.logger.Info("Starting price tracking for all items")

	for _, item := range t.config.Items {
		if err := t.TrackItem(ctx, item); err != nil {
			t.logger.Error("Failed to track item", "item", item.ID, "error", err)
			continue
		}
	}

	t.logger.Info("Price tracking completed")
	return nil
}

func (t *Tracker) CheckAlerts(ctx context.Context) error {
	t.logger.Info("Checking price alerts")

	for _, item := range t.config.Items {
		if err := t.checkItemAlerts(ctx, item); err != nil {
			t.logger.Error("Failed to check alerts for item", "item", item.ID, "error", err)
			continue
		}
	}

	return nil
}

func (t *Tracker) checkItemAlerts(ctx context.Context, item config.ItemConfig) error {
	// Get latest price
	latest, err := t.storage.GetLatestPrice(ctx, item.ID)
	if err != nil {
		return fmt.Errorf("failed to get latest price: %w", err)
	}
	if latest == nil {
		return nil // No price data yet
	}

	// Get recent prices for comparison
	prices, err := t.storage.GetPrices(ctx, item.ID, 5)
	if err != nil {
		return fmt.Errorf("failed to get price history: %w", err)
	}
	if len(prices) < 2 {
		return nil // Need at least 2 prices for comparison
	}

	// Check target price alert
	if item.TargetPrice != nil && latest.Price <= *item.TargetPrice {
		t.logger.Info("Target price reached", 
			"item", item.ID, 
			"current", latest.Price, 
			"target", *item.TargetPrice,
		)
		// TODO: Send notification
	}

	// Check percent drop alert
	percentDrop := item.PercentDrop
	if percentDrop == nil {
		percentDrop = &t.config.Rules.PercentDrop
	}

	if *percentDrop > 0 {
		previousPrice := prices[1].Price
		dropPercent := ((previousPrice - latest.Price) / previousPrice) * 100
		
		if dropPercent >= *percentDrop {
			t.logger.Info("Price drop alert", 
				"item", item.ID, 
				"current", latest.Price, 
				"previous", previousPrice,
				"drop_percent", dropPercent,
			)
			// TODO: Send notification
		}
	}

	return nil
}