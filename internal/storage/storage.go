package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/makalin/pricetrek/internal/config"
)

type Storage interface {
	Init() error
	Close() error
	SavePrice(ctx context.Context, itemID string, price float64, currency string, meta map[string]interface{}) error
	GetPrices(ctx context.Context, itemID string, limit int) ([]PriceSample, error)
	GetLatestPrice(ctx context.Context, itemID string) (*PriceSample, error)
	GetItems(ctx context.Context) ([]Item, error)
	SaveItem(ctx context.Context, item Item) error
	DeleteItem(ctx context.Context, itemID string) error
	GetItem(ctx context.Context, itemID string) (*Item, error)
}

type PriceSample struct {
	ItemID   string                 `json:"item_id"`
	Time     time.Time              `json:"time"`
	Price    float64                `json:"price"`
	Currency string                 `json:"currency"`
	Meta     map[string]interface{} `json:"meta,omitempty"`
}

type Item struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Provider    string   `json:"provider"`
	Selector    string   `json:"selector"`
	Currency    string   `json:"currency"`
	TargetPrice *float64 `json:"target_price,omitempty"`
	PercentDrop *float64 `json:"percent_drop,omitempty"`
	Schedule    string   `json:"schedule"`
	Regex       string   `json:"regex,omitempty"`
	Attr        string   `json:"attr,omitempty"`
	Command     string   `json:"command,omitempty"`
}

type sqliteStorage struct {
	db *sql.DB
}

func New(cfg config.StorageConfig) (Storage, error) {
	// Ensure directory exists
	dir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	db, err := sql.Open("sqlite3", cfg.Path+"?_journal_mode=WAL&_synchronous=NORMAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &sqliteStorage{db: db}, nil
}

func (s *sqliteStorage) Init() error {
	// Create prices table
	createPricesTable := `
	CREATE TABLE IF NOT EXISTS prices (
		item_id TEXT NOT NULL,
		ts DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		price REAL NOT NULL,
		currency TEXT NOT NULL,
		meta TEXT
	);
	`
	if _, err := s.db.Exec(createPricesTable); err != nil {
		return fmt.Errorf("failed to create prices table: %w", err)
	}

	// Create index
	createIndex := `
	CREATE INDEX IF NOT EXISTS idx_prices_item_ts ON prices(item_id, ts DESC);
	`
	if _, err := s.db.Exec(createIndex); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// Create items table
	createItemsTable := `
	CREATE TABLE IF NOT EXISTS items (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		provider TEXT NOT NULL,
		selector TEXT,
		currency TEXT NOT NULL,
		target_price REAL,
		percent_drop REAL,
		schedule TEXT,
		regex TEXT,
		attr TEXT,
		command TEXT
	);
	`
	if _, err := s.db.Exec(createItemsTable); err != nil {
		return fmt.Errorf("failed to create items table: %w", err)
	}

	return nil
}

func (s *sqliteStorage) Close() error {
	return s.db.Close()
}

func (s *sqliteStorage) SavePrice(ctx context.Context, itemID string, price float64, currency string, meta map[string]interface{}) error {
	var metaJSON string
	if meta != nil {
		metaBytes, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("failed to marshal meta: %w", err)
		}
		metaJSON = string(metaBytes)
	}

	query := `
	INSERT INTO prices (item_id, ts, price, currency, meta)
	VALUES (?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query, itemID, time.Now(), price, currency, metaJSON)
	if err != nil {
		return fmt.Errorf("failed to save price: %w", err)
	}

	return nil
}

func (s *sqliteStorage) GetPrices(ctx context.Context, itemID string, limit int) ([]PriceSample, error) {
	query := `
	SELECT item_id, ts, price, currency, meta
	FROM prices
	WHERE item_id = ?
	ORDER BY ts DESC
	LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, itemID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query prices: %w", err)
	}
	defer rows.Close()

	var samples []PriceSample
	for rows.Next() {
		var sample PriceSample
		var metaJSON sql.NullString

		err := rows.Scan(&sample.ItemID, &sample.Time, &sample.Price, &sample.Currency, &metaJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan price: %w", err)
		}

		if metaJSON.Valid && metaJSON.String != "" {
			if err := json.Unmarshal([]byte(metaJSON.String), &sample.Meta); err != nil {
				return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
			}
		}

		samples = append(samples, sample)
	}

	return samples, nil
}

func (s *sqliteStorage) GetLatestPrice(ctx context.Context, itemID string) (*PriceSample, error) {
	query := `
	SELECT item_id, ts, price, currency, meta
	FROM prices
	WHERE item_id = ?
	ORDER BY ts DESC
	LIMIT 1
	`

	var sample PriceSample
	var metaJSON sql.NullString

	err := s.db.QueryRowContext(ctx, query, itemID).Scan(
		&sample.ItemID, &sample.Time, &sample.Price, &sample.Currency, &metaJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest price: %w", err)
	}

	if metaJSON.Valid && metaJSON.String != "" {
		if err := json.Unmarshal([]byte(metaJSON.String), &sample.Meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
		}
	}

	return &sample, nil
}

func (s *sqliteStorage) GetItems(ctx context.Context) ([]Item, error) {
	query := `
	SELECT id, name, url, provider, selector, currency, target_price, percent_drop, schedule, regex, attr, command
	FROM items
	ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var targetPrice, percentDrop sql.NullFloat64

		err := rows.Scan(
			&item.ID, &item.Name, &item.URL, &item.Provider, &item.Selector,
			&item.Currency, &targetPrice, &percentDrop, &item.Schedule,
			&item.Regex, &item.Attr, &item.Command,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}

		if targetPrice.Valid {
			item.TargetPrice = &targetPrice.Float64
		}
		if percentDrop.Valid {
			item.PercentDrop = &percentDrop.Float64
		}

		items = append(items, item)
	}

	return items, nil
}

func (s *sqliteStorage) SaveItem(ctx context.Context, item Item) error {
	query := `
	INSERT OR REPLACE INTO items (id, name, url, provider, selector, currency, target_price, percent_drop, schedule, regex, attr, command)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		item.ID, item.Name, item.URL, item.Provider, item.Selector,
		item.Currency, item.TargetPrice, item.PercentDrop, item.Schedule,
		item.Regex, item.Attr, item.Command,
	)
	if err != nil {
		return fmt.Errorf("failed to save item: %w", err)
	}

	return nil
}

func (s *sqliteStorage) DeleteItem(ctx context.Context, itemID string) error {
	// Delete item
	query := `DELETE FROM items WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, itemID)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	// Delete associated prices
	query = `DELETE FROM prices WHERE item_id = ?`
	_, err = s.db.ExecContext(ctx, query, itemID)
	if err != nil {
		return fmt.Errorf("failed to delete prices: %w", err)
	}

	return nil
}

func (s *sqliteStorage) GetItem(ctx context.Context, itemID string) (*Item, error) {
	query := `
	SELECT id, name, url, provider, selector, currency, target_price, percent_drop, schedule, regex, attr, command
	FROM items
	WHERE id = ?
	`

	var item Item
	var targetPrice, percentDrop sql.NullFloat64

	err := s.db.QueryRowContext(ctx, query, itemID).Scan(
		&item.ID, &item.Name, &item.URL, &item.Provider, &item.Selector,
		&item.Currency, &targetPrice, &percentDrop, &item.Schedule,
		&item.Regex, &item.Attr, &item.Command,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	if targetPrice.Valid {
		item.TargetPrice = &targetPrice.Float64
	}
	if percentDrop.Valid {
		item.PercentDrop = &percentDrop.Float64
	}

	return &item, nil
}