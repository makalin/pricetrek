package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/makalin/pricetrek/internal/storage"
)

// ExportItems exports items to CSV format
func ExportItems(items []storage.Item, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"id", "name", "url", "provider", "selector", "currency",
		"target_price", "percent_drop", "schedule", "regex", "attr", "command",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write items
	for _, item := range items {
		record := []string{
			item.ID,
			item.Name,
			item.URL,
			item.Provider,
			item.Selector,
			item.Currency,
		}

		// Handle optional fields
		if item.TargetPrice != nil {
			record = append(record, fmt.Sprintf("%.2f", *item.TargetPrice))
		} else {
			record = append(record, "")
		}

		if item.PercentDrop != nil {
			record = append(record, fmt.Sprintf("%.2f", *item.PercentDrop))
		} else {
			record = append(record, "")
		}

		record = append(record, item.Schedule, item.Regex, item.Attr, item.Command)

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

// ImportItems imports items from CSV format
func ImportItems(filename string) ([]storage.Item, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least a header and one data row")
	}

	// Skip header row
	records = records[1:]

	var items []storage.Item
	for i, record := range records {
		if len(record) < 6 {
			return nil, fmt.Errorf("row %d has insufficient columns", i+2)
		}

		item := storage.Item{
			ID:       record[0],
			Name:     record[1],
			URL:      record[2],
			Provider: record[3],
			Selector: record[4],
			Currency: record[5],
		}

		// Parse optional fields
		if len(record) > 6 && record[6] != "" {
			if price, err := strconv.ParseFloat(record[6], 64); err == nil {
				item.TargetPrice = &price
			}
		}

		if len(record) > 7 && record[7] != "" {
			if percent, err := strconv.ParseFloat(record[7], 64); err == nil {
				item.PercentDrop = &percent
			}
		}

		if len(record) > 8 {
			item.Schedule = record[8]
		}
		if len(record) > 9 {
			item.Regex = record[9]
		}
		if len(record) > 10 {
			item.Attr = record[10]
		}
		if len(record) > 11 {
			item.Command = record[11]
		}

		items = append(items, item)
	}

	return items, nil
}

// ExportPrices exports price history to CSV format
func ExportPrices(prices []storage.PriceSample, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"item_id", "timestamp", "price", "currency", "in_stock"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write prices
	for _, price := range prices {
		inStock := "true"
		if price.Meta != nil {
			if stock, ok := price.Meta["in_stock"].(bool); ok && !stock {
				inStock = "false"
			}
		}

		record := []string{
			price.ItemID,
			price.Time.Format(time.RFC3339),
			fmt.Sprintf("%.2f", price.Price),
			price.Currency,
			inStock,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}