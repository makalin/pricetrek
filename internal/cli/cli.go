package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/makalin/pricetrek/internal/config"
	"github.com/makalin/pricetrek/internal/csv"
	"github.com/makalin/pricetrek/internal/logger"
	"github.com/makalin/pricetrek/internal/providers"
	"github.com/makalin/pricetrek/internal/scheduler"
	"github.com/makalin/pricetrek/internal/storage"
	"github.com/makalin/pricetrek/internal/tools"
	"github.com/makalin/pricetrek/internal/tracker"
	"github.com/makalin/pricetrek/internal/utils"
)

type CLI struct {
	config *config.Config
	logger *logger.Logger
	storage storage.Storage
	tracker *tracker.Tracker
}

func New(cfg *config.Config, log *logger.Logger) *CLI {
	return &CLI{
		config: cfg,
		logger: log,
	}
}

func (c *CLI) Execute(ctx context.Context, args []string) error {
	command := args[0]
	
	// Handle init command without requiring config
	if command == "init" {
		return c.handleInit(args[1:])
	}
	
	// Initialize storage
	var err error
	c.storage, err = storage.New(c.config.Storage)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer c.storage.Close()

	// Initialize tracker
	c.tracker = tracker.New(c.config, c.storage, c.logger)

	switch command {
	case "add":
		return c.handleAdd(args[1:])
	case "rm", "remove":
		return c.handleRemove(args[1:])
	case "ls", "list":
		return c.handleList(args[1:])
	case "show":
		return c.handleShow(args[1:])
	case "track":
		return c.handleTrack(ctx, args[1:])
	case "alert":
		return c.handleAlert(ctx, args[1:])
	case "export":
		return c.handleExport(args[1:])
	case "import":
		return c.handleImport(args[1:])
	case "doctor":
		return c.handleDoctor(args[1:])
	case "schedule":
		return c.handleSchedule(args[1:])
	case "backup":
		return c.handleBackup(args[1:])
	case "restore":
		return c.handleRestore(args[1:])
	case "monitor":
		return c.handleMonitor(args[1:])
	case "help", "-h", "--help":
		c.Help()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func (c *CLI) Help() {
	fmt.Fprintf(os.Stderr, `PriceTrek - A tiny, fast terminal agent to track product prices

USAGE:
    pricetrek <command> [options]

COMMANDS:
    init                       Scaffold config & DB
    add --name --url ...       Add a product (or use --from yaml/csv)
    rm <id>                    Remove item
    ls [--json]                List watchlist
    show <id> [--spark]        Price history with sparkline
    track [--once|--loop]      Run trackers (respects per-item schedule)
    alert --dry-run            Re-evaluate rules & send alerts
    export --csv out.csv       Dump history
    import --csv in.csv        Import items
    doctor                     Env & provider health check
    schedule --hourly|--daily  Print OS-specific scheduler instructions
    backup --output file       Create backup
    restore --file backup      Restore backup
    monitor [--once]           System monitoring
    help                       Show this help message

OPTIONS:
    --config string    Path to configuration file (default "pricetrek.yaml")
    --verbose          Enable verbose logging
    --version          Show version information

EXAMPLES:
    # Initialize workspace
    pricetrek init

    # Add a product
    pricetrek add --name "Samsung 990 Pro 2TB" --url "https://example.com" --provider generic --selector ".price .value" --currency TRY --target 4250

    # Run tracking once
    pricetrek track --once --verbose

    # Show price history with sparkline
    pricetrek show 990pro-2tb --spark

    # Export data
    pricetrek export --csv history.csv

    # Create backup
    pricetrek backup --output backup.tar.gz

    # Monitor system
    pricetrek monitor --once

For more information, visit: https://github.com/makalin/pricetrek
`)
}

func (c *CLI) handleInit(args []string) error {
	c.logger.Info("Initializing PriceTrek workspace")
	
	// Create data directory
	if err := os.MkdirAll("data", 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create default config if it doesn't exist
	configPath := "pricetrek.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := &config.Config{
			Storage: config.StorageConfig{
				Driver: "sqlite",
				Path:   "./data/trek.db",
			},
			Defaults: config.DefaultsConfig{
				Currency:  "USD",
				Timezone:  "UTC",
				UserAgent: "PriceTrek/0.1 (+https://github.com/makalin/pricetrek)",
				Retry: config.RetryConfig{
					Attempts:  3,
					BaseDelay: 800,
					MaxDelay:  7000,
				},
				HTTPTimeout: 20,
				CacheTTL:    30,
				Headless: config.HeadlessConfig{
					Enabled:   false,
					WaitUntil: "networkidle",
				},
			},
			Rules: config.RulesConfig{
				PercentDrop: 8.0,
			},
			Items: []config.ItemConfig{},
		}

		if err := defaultConfig.Save(configPath); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
		c.logger.Info("Created default configuration", "file", configPath)
	}

	// Initialize storage with default config
	storageConfig := config.StorageConfig{
		Driver: "sqlite",
		Path:   "./data/trek.db",
	}
	
	storage, err := storage.New(storageConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer storage.Close()

	if err := storage.Init(); err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	c.logger.Info("PriceTrek workspace initialized successfully")
	return nil
}

func (c *CLI) handleAdd(args []string) error {
	var (
		name     = flag.String("name", "", "Product name")
		url      = flag.String("url", "", "Product URL")
		provider = flag.String("provider", "generic", "Provider type (generic, exec)")
		selector = flag.String("selector", "", "CSS selector for price extraction")
		currency = flag.String("currency", "", "Currency code (USD, EUR, TRY, etc.)")
		target   = flag.Float64("target", 0, "Target price")
		percent  = flag.Float64("percent", 0, "Percent drop threshold")
		schedule = flag.String("schedule", "hourly", "Schedule (hourly, daily, cron)")
		regex    = flag.String("regex", "", "Regex pattern for price cleanup")
		attr     = flag.String("attr", "", "Attribute to extract (text, content, data-price)")
		command  = flag.String("command", "", "Command for exec provider")
		fromFile = flag.String("from", "", "Import from file (yaml, csv)")
		jsonFlag = flag.Bool("json", false, "Output in JSON format")
	)

	// Parse flags
	flag.CommandLine.Parse(args)

	// Handle import from file
	if *fromFile != "" {
		return c.handleImportFromFile(*fromFile, *jsonFlag)
	}

	// Validate required fields
	if *name == "" {
		return fmt.Errorf("name is required")
	}
	if *url == "" {
		return fmt.Errorf("url is required")
	}
	if *provider == "generic" && *selector == "" {
		return fmt.Errorf("selector is required for generic provider")
	}
	if *provider == "exec" && *command == "" {
		return fmt.Errorf("command is required for exec provider")
	}

	// Use defaults from config
	if *currency == "" {
		*currency = c.config.Defaults.Currency
	}
	if *percent == 0 {
		*percent = c.config.Rules.PercentDrop
	}

	// Create item
	item := storage.Item{
		ID:          c.generateItemID(*name),
		Name:        *name,
		URL:         *url,
		Provider:    *provider,
		Selector:    *selector,
		Currency:    *currency,
		Schedule:    *schedule,
		Regex:       *regex,
		Attr:        *attr,
		Command:     *command,
	}

	if *target > 0 {
		item.TargetPrice = target
	}
	if *percent > 0 {
		item.PercentDrop = percent
	}

	// Save item
	ctx := context.Background()
	if err := c.storage.SaveItem(ctx, item); err != nil {
		return fmt.Errorf("failed to save item: %w", err)
	}

	if *jsonFlag {
		// Output JSON
		jsonData, err := json.MarshalIndent(item, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	} else {
		c.logger.Info("Item added successfully", 
			"id", item.ID, 
			"name", item.Name,
			"url", item.URL,
			"provider", item.Provider,
		)
	}

	return nil
}

func (c *CLI) generateItemID(name string) string {
	// Simple ID generation based on name
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "_", "-")
	
	// Remove special characters
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	id = reg.ReplaceAllString(id, "")
	
	// Ensure it's not empty
	if id == "" {
		id = "item"
	}
	
	// Add timestamp to make it unique
	return fmt.Sprintf("%s-%d", id, time.Now().Unix())
}

func (c *CLI) handleImportFromFile(filename string, jsonOutput bool) error {
	// TODO: Implement file import
	return fmt.Errorf("file import not implemented yet")
}

func (c *CLI) handleRemove(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("item ID is required")
	}

	itemID := args[0]
	
	// Confirm deletion
	fmt.Printf("Are you sure you want to delete item '%s'? (y/N): ", itemID)
	var response string
	fmt.Scanln(&response)
	
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		c.logger.Info("Deletion cancelled")
		return nil
	}

	// Delete item
	ctx := context.Background()
	if err := c.storage.DeleteItem(ctx, itemID); err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	c.logger.Info("Item deleted successfully", "id", itemID)
	return nil
}

func (c *CLI) handleList(args []string) error {
	var (
		jsonFlag = flag.Bool("json", false, "Output in JSON format")
		verbose  = flag.Bool("verbose", false, "Show detailed information")
	)

	// Parse flags
	flag.CommandLine.Parse(args)

	// Get all items
	ctx := context.Background()
	items, err := c.storage.GetItems(ctx)
	if err != nil {
		return fmt.Errorf("failed to get items: %w", err)
	}

	if len(items) == 0 {
		c.logger.Info("No items found")
		return nil
	}

	if *jsonFlag {
		// Output JSON
		jsonData, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	} else {
		// Output table format
		c.printItemsTable(items, *verbose)
	}

	return nil
}

func (c *CLI) printItemsTable(items []storage.Item, verbose bool) {
	// Print header
	fmt.Printf("%-20s %-30s %-15s %-10s %-10s %-10s\n", 
		"ID", "Name", "Provider", "Currency", "Target", "Schedule")
	fmt.Println(strings.Repeat("-", 95))

	// Print items
	for _, item := range items {
		target := "-"
		if item.TargetPrice != nil {
			target = fmt.Sprintf("%.2f", *item.TargetPrice)
		}

		fmt.Printf("%-20s %-30s %-15s %-10s %-10s %-10s\n",
			item.ID,
			truncateString(item.Name, 30),
			item.Provider,
			item.Currency,
			target,
			item.Schedule,
		)

		if verbose {
			fmt.Printf("  URL: %s\n", item.URL)
			if item.Selector != "" {
				fmt.Printf("  Selector: %s\n", item.Selector)
			}
			if item.Regex != "" {
				fmt.Printf("  Regex: %s\n", item.Regex)
			}
			if item.Attr != "" {
				fmt.Printf("  Attribute: %s\n", item.Attr)
			}
			if item.Command != "" {
				fmt.Printf("  Command: %s\n", item.Command)
			}
			if item.PercentDrop != nil {
				fmt.Printf("  Percent Drop: %.1f%%\n", *item.PercentDrop)
			}
			fmt.Println()
		}
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (c *CLI) handleShow(args []string) error {
	var (
		sparkFlag = flag.Bool("spark", false, "Show sparkline")
		limit     = flag.Int("limit", 30, "Number of price points to show")
		jsonFlag  = flag.Bool("json", false, "Output in JSON format")
	)

	// Parse flags
	flag.CommandLine.Parse(args)

	if len(args) == 0 {
		return fmt.Errorf("item ID is required")
	}

	itemID := args[0]

	// Get item
	ctx := context.Background()
	item, err := c.storage.GetItem(ctx, itemID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}
	if item == nil {
		return fmt.Errorf("item not found: %s", itemID)
	}

	// Get price history
	prices, err := c.storage.GetPrices(ctx, itemID, *limit)
	if err != nil {
		return fmt.Errorf("failed to get price history: %w", err)
	}

	if len(prices) == 0 {
		c.logger.Info("No price history found for item", "id", itemID)
		return nil
	}

	if *jsonFlag {
		// Output JSON
		response := map[string]interface{}{
			"item":   item,
			"prices": prices,
		}
		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	} else {
		// Output formatted display
		c.printItemDetails(item, prices, *sparkFlag)
	}

	return nil
}

func (c *CLI) printItemDetails(item *storage.Item, prices []storage.PriceSample, showSparkline bool) {
	fmt.Printf("Item: %s (%s)\n", item.Name, item.ID)
	fmt.Printf("URL: %s\n", item.URL)
	fmt.Printf("Provider: %s\n", item.Provider)
	fmt.Printf("Currency: %s\n", item.Currency)
	fmt.Printf("Schedule: %s\n", item.Schedule)
	
	if item.TargetPrice != nil {
		fmt.Printf("Target Price: %.2f %s\n", *item.TargetPrice, item.Currency)
	}
	if item.PercentDrop != nil {
		fmt.Printf("Percent Drop Alert: %.1f%%\n", *item.PercentDrop)
	}
	
	fmt.Println()

	// Show price history
	fmt.Printf("Price History (%d points):\n", len(prices))
	fmt.Println(strings.Repeat("-", 50))

	// Extract price values for sparkline
	priceValues := make([]float64, len(prices))
	for i, price := range prices {
		priceValues[i] = price.Price
	}

	// Show sparkline if requested
	if showSparkline && len(priceValues) > 1 {
		sparkline := utils.GenerateSparkline(priceValues, 50)
		fmt.Printf("Price Trend: %s\n", sparkline)
		fmt.Println()
	}

	// Show recent prices
	for i, price := range prices {
		if i >= 10 { // Show only last 10
			break
		}
		formattedPrice := utils.FormatPrice(price.Price, price.Currency)
		fmt.Printf("%s: %s", price.Time.Format("2006-01-02 15:04:05"), formattedPrice)
		
		// Show price change if we have previous price
		if i < len(prices)-1 {
			prevPrice := prices[i+1].Price
			change := utils.CalculatePriceChange(prevPrice, price.Price)
			if change > 0 {
				fmt.Printf(" (+%.1f%%)", change)
			} else if change < 0 {
				fmt.Printf(" (%.1f%%)", change)
			}
		}
		fmt.Println()
	}

	// Show statistics
	if len(priceValues) > 1 {
		min, max, avg, median := utils.CalculateStats(priceValues)
		fmt.Println()
		fmt.Printf("Statistics:\n")
		fmt.Printf("  Min: %s\n", utils.FormatPrice(min, item.Currency))
		fmt.Printf("  Max: %s\n", utils.FormatPrice(max, item.Currency))
		fmt.Printf("  Avg: %s\n", utils.FormatPrice(avg, item.Currency))
		fmt.Printf("  Median: %s\n", utils.FormatPrice(median, item.Currency))
	}
}

func (c *CLI) handleTrack(ctx context.Context, args []string) error {
	var (
		onceFlag     = flag.Bool("once", false, "Run tracking once")
		loopFlag     = flag.Bool("loop", false, "Run tracking in a loop")
		itemID       = flag.String("id", "", "Track specific item ID")
		noCacheFlag  = flag.Bool("no-cache", false, "Disable caching")
		respectCache = flag.Bool("respect-cache", false, "Respect cache TTL")
		interval     = flag.Duration("interval", 1*time.Hour, "Loop interval")
	)

	// Parse flags
	flag.CommandLine.Parse(args)

	// Determine mode
	if *onceFlag && *loopFlag {
		return fmt.Errorf("cannot specify both --once and --loop")
	}
	if !*onceFlag && !*loopFlag {
		*onceFlag = true // Default to once
	}

	if *onceFlag {
		return c.trackOnce(ctx, *itemID, *noCacheFlag, *respectCache)
	} else {
		return c.trackLoop(ctx, *itemID, *noCacheFlag, *respectCache, *interval)
	}
}

func (c *CLI) trackOnce(ctx context.Context, itemID string, noCache, respectCache bool) error {
	c.logger.Info("Starting one-time price tracking")

	if itemID != "" {
		// Track specific item
		item, err := c.storage.GetItem(ctx, itemID)
		if err != nil {
			return fmt.Errorf("failed to get item: %w", err)
		}
		if item == nil {
			return fmt.Errorf("item not found: %s", itemID)
		}

		// Convert to config format
		itemConfig := config.ItemConfig{
			ID:          item.ID,
			Name:        item.Name,
			URL:         item.URL,
			Provider:    item.Provider,
			Selector:    item.Selector,
			Currency:    item.Currency,
			TargetPrice: item.TargetPrice,
			PercentDrop: item.PercentDrop,
			Schedule:    item.Schedule,
			Regex:       item.Regex,
			Attr:        item.Attr,
			Command:     item.Command,
		}

		return c.tracker.TrackItem(ctx, itemConfig)
	} else {
		// Track all items
		return c.tracker.TrackAll(ctx)
	}
}

func (c *CLI) trackLoop(ctx context.Context, itemID string, noCache, respectCache bool, interval time.Duration) error {
	c.logger.Info("Starting continuous price tracking", "interval", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Tracking stopped")
			return ctx.Err()
		case <-ticker.C:
			c.logger.Info("Running scheduled tracking")
			if err := c.trackOnce(ctx, itemID, noCache, respectCache); err != nil {
				c.logger.Error("Tracking failed", "error", err)
			}
		}
	}
}

func (c *CLI) handleAlert(ctx context.Context, args []string) error {
	// TODO: Implement alert command
	c.logger.Info("Alert command not yet implemented")
	return nil
}

func (c *CLI) handleExport(args []string) error {
	var (
		csvFlag    = flag.String("csv", "", "Export to CSV file")
		itemsFlag  = flag.Bool("items", false, "Export items")
		pricesFlag = flag.Bool("prices", false, "Export price history")
		itemID     = flag.String("id", "", "Export specific item")
	)

	// Parse flags
	flag.CommandLine.Parse(args)

	if *csvFlag == "" {
		return fmt.Errorf("CSV filename is required (--csv)")
	}

	if !*itemsFlag && !*pricesFlag {
		*itemsFlag = true // Default to items
	}

	ctx := context.Background()

	if *itemsFlag {
		// Export items
		items, err := c.storage.GetItems(ctx)
		if err != nil {
			return fmt.Errorf("failed to get items: %w", err)
		}

		if err := csv.ExportItems(items, *csvFlag); err != nil {
			return fmt.Errorf("failed to export items: %w", err)
		}

		c.logger.Info("Items exported successfully", "file", *csvFlag, "count", len(items))
	}

	if *pricesFlag {
		// Export prices
		if *itemID != "" {
			// Export specific item's prices
			prices, err := c.storage.GetPrices(ctx, *itemID, 1000) // Get up to 1000 prices
			if err != nil {
				return fmt.Errorf("failed to get prices: %w", err)
			}

			if err := csv.ExportPrices(prices, *csvFlag); err != nil {
				return fmt.Errorf("failed to export prices: %w", err)
			}

			c.logger.Info("Prices exported successfully", "file", *csvFlag, "item", *itemID, "count", len(prices))
		} else {
			// Export all prices
			items, err := c.storage.GetItems(ctx)
			if err != nil {
				return fmt.Errorf("failed to get items: %w", err)
			}

			var allPrices []storage.PriceSample
			for _, item := range items {
				prices, err := c.storage.GetPrices(ctx, item.ID, 1000)
				if err != nil {
					c.logger.Error("Failed to get prices for item", "item", item.ID, "error", err)
					continue
				}
				allPrices = append(allPrices, prices...)
			}

			if err := csv.ExportPrices(allPrices, *csvFlag); err != nil {
				return fmt.Errorf("failed to export prices: %w", err)
			}

			c.logger.Info("All prices exported successfully", "file", *csvFlag, "count", len(allPrices))
		}
	}

	return nil
}

func (c *CLI) handleImport(args []string) error {
	var (
		csvFlag = flag.String("csv", "", "Import from CSV file")
		yamlFlag = flag.String("yaml", "", "Import from YAML file")
	)

	// Parse flags
	flag.CommandLine.Parse(args)

	if *csvFlag == "" && *yamlFlag == "" {
		return fmt.Errorf("import file is required (--csv or --yaml)")
	}

	ctx := context.Background()

	if *csvFlag != "" {
		// Import from CSV
		items, err := csv.ImportItems(*csvFlag)
		if err != nil {
			return fmt.Errorf("failed to import CSV: %w", err)
		}

		// Save items to storage
		for _, item := range items {
			if err := c.storage.SaveItem(ctx, item); err != nil {
				c.logger.Error("Failed to save item", "item", item.ID, "error", err)
				continue
			}
		}

		c.logger.Info("CSV import completed", "file", *csvFlag, "count", len(items))
	}

	if *yamlFlag != "" {
		// Import from YAML
		cfg, err := config.Load(*yamlFlag)
		if err != nil {
			return fmt.Errorf("failed to load YAML: %w", err)
		}

		// Convert config items to storage items
		for _, itemConfig := range cfg.Items {
			item := storage.Item{
				ID:          itemConfig.ID,
				Name:        itemConfig.Name,
				URL:         itemConfig.URL,
				Provider:    itemConfig.Provider,
				Selector:    itemConfig.Selector,
				Currency:    itemConfig.Currency,
				TargetPrice: itemConfig.TargetPrice,
				PercentDrop: itemConfig.PercentDrop,
				Schedule:    itemConfig.Schedule,
				Regex:       itemConfig.Regex,
				Attr:        itemConfig.Attr,
				Command:     itemConfig.Command,
			}

			if err := c.storage.SaveItem(ctx, item); err != nil {
				c.logger.Error("Failed to save item", "item", item.ID, "error", err)
				continue
			}
		}

		c.logger.Info("YAML import completed", "file", *yamlFlag, "count", len(cfg.Items))
	}

	return nil
}

func (c *CLI) handleDoctor(args []string) error {
	c.logger.Info("Running PriceTrek health check...")
	
	var issues []string
	
	// Check database connection
	if err := c.checkDatabase(); err != nil {
		issues = append(issues, fmt.Sprintf("Database: %v", err))
	} else {
		c.logger.Info("✓ Database connection OK")
	}
	
	// Check network connectivity
	if err := c.checkNetwork(); err != nil {
		issues = append(issues, fmt.Sprintf("Network: %v", err))
	} else {
		c.logger.Info("✓ Network connectivity OK")
	}
	
	// Check providers
	if err := c.checkProviders(); err != nil {
		issues = append(issues, fmt.Sprintf("Providers: %v", err))
	} else {
		c.logger.Info("✓ Providers OK")
	}
	
	// Check notifications
	if err := c.checkNotifications(); err != nil {
		issues = append(issues, fmt.Sprintf("Notifications: %v", err))
	} else {
		c.logger.Info("✓ Notifications OK")
	}
	
	// Check configuration
	if err := c.checkConfiguration(); err != nil {
		issues = append(issues, fmt.Sprintf("Configuration: %v", err))
	} else {
		c.logger.Info("✓ Configuration OK")
	}
	
	if len(issues) > 0 {
		c.logger.Error("Health check found issues:")
		for _, issue := range issues {
			c.logger.Error("  - " + issue)
		}
		return fmt.Errorf("health check failed with %d issues", len(issues))
	}
	
	c.logger.Info("✓ All systems healthy!")
	return nil
}

func (c *CLI) checkDatabase() error {
	ctx := context.Background()
	_, err := c.storage.GetItems(ctx)
	return err
}

func (c *CLI) checkNetwork() error {
	// Simple network check by trying to connect to a reliable endpoint
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://httpbin.org/status/200")
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *CLI) checkProviders() error {
	// Test generic provider
	provider, err := providers.GetProvider("generic", c.config.Defaults)
	if err != nil {
		return err
	}
	
	// Test with a simple URL
	testItem := config.ItemConfig{
		URL:      "https://httpbin.org/html",
		Provider: "generic",
		Selector: "h1",
		Currency: "USD",
	}
	
	ctx := context.Background()
	_, err = provider.Fetch(ctx, testItem)
	return err
}

func (c *CLI) checkNotifications() error {
	// Check if notification services are properly configured
	if c.config.Notifications.Email.Enabled && c.config.Notifications.Email.From == "" {
		return fmt.Errorf("email notifications enabled but no 'from' address configured")
	}
	if c.config.Notifications.Telegram.Enabled && c.config.Notifications.Telegram.ChatID == "" {
		return fmt.Errorf("telegram notifications enabled but no chat ID configured")
	}
	if c.config.Notifications.Slack.Enabled && c.config.Notifications.Slack.Webhook == "" {
		return fmt.Errorf("slack notifications enabled but no webhook configured")
	}
	return nil
}

func (c *CLI) checkConfiguration() error {
	// Check for required configuration values
	if c.config.Defaults.Currency == "" {
		return fmt.Errorf("default currency not set")
	}
	if c.config.Storage.Driver == "" {
		return fmt.Errorf("storage driver not set")
	}
	return nil
}

func (c *CLI) handleSchedule(args []string) error {
	var (
		hourlyFlag = flag.Bool("hourly", false, "Generate hourly schedule")
		dailyFlag  = flag.Bool("daily", false, "Generate daily schedule")
	)

	// Parse flags
	flag.CommandLine.Parse(args)

	if *hourlyFlag && *dailyFlag {
		return fmt.Errorf("cannot specify both --hourly and --daily")
	}
	if !*hourlyFlag && !*dailyFlag {
		*hourlyFlag = true // Default to hourly
	}

	sched := scheduler.New()

	if *hourlyFlag {
		schedule, err := sched.GenerateHourlySchedule()
		if err != nil {
			return fmt.Errorf("failed to generate hourly schedule: %w", err)
		}
		fmt.Println(schedule)
	} else {
		schedule, err := sched.GenerateDailySchedule()
		if err != nil {
			return fmt.Errorf("failed to generate daily schedule: %w", err)
		}
		fmt.Println(schedule)
	}

	return nil
}

func (c *CLI) handleBackup(args []string) error {
	var (
		outputFlag = flag.String("output", "", "Backup output file")
		dirFlag    = flag.String("dir", "./backups", "Backup directory")
	)

	// Parse flags
	flag.CommandLine.Parse(args)

	backupManager := tools.NewBackupManager(*dirFlag)

	// Create backup
	backupFile, err := backupManager.CreateBackup("./data")
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	if *outputFlag != "" {
		// Move backup to specified location
		if err := os.Rename(backupFile, *outputFlag); err != nil {
			return fmt.Errorf("failed to move backup file: %w", err)
		}
		backupFile = *outputFlag
	}

	c.logger.Info("Backup created successfully", "file", backupFile)
	return nil
}

func (c *CLI) handleRestore(args []string) error {
	var (
		backupFile = flag.String("file", "", "Backup file to restore")
		targetDir  = flag.String("target", "./data", "Target directory")
	)

	// Parse flags
	flag.CommandLine.Parse(args)

	if *backupFile == "" {
		return fmt.Errorf("backup file is required (--file)")
	}

	backupManager := tools.NewBackupManager("./backups")

	// Restore backup
	if err := backupManager.RestoreBackup(*backupFile, *targetDir); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	c.logger.Info("Backup restored successfully", "file", *backupFile, "target", *targetDir)
	return nil
}

func (c *CLI) handleMonitor(args []string) error {
	var (
		intervalFlag = flag.Duration("interval", 5*time.Second, "Monitoring interval")
		onceFlag     = flag.Bool("once", false, "Show stats once and exit")
	)

	// Parse flags
	flag.CommandLine.Parse(args)

	monitor := tools.NewSystemMonitor()

	if *onceFlag {
		// Show stats once
		stats := monitor.GetSystemStats()
		c.printSystemStats(stats)
		return nil
	}

	// Continuous monitoring
	c.logger.Info("Starting system monitoring", "interval", *intervalFlag)
	ctx := context.Background()

	monitor.MonitorLoop(ctx, *intervalFlag, func(stats tools.SystemStats) {
		c.printSystemStats(stats)
	})

	return nil
}

func (c *CLI) printSystemStats(stats tools.SystemStats) {
	fmt.Printf("\n=== System Statistics ===\n")
	fmt.Printf("Uptime: %v\n", stats.Uptime)
	fmt.Printf("Go Routines: %d\n", stats.GoRoutines)
	fmt.Printf("Memory Allocated: %s\n", stats.FormatBytes(stats.MemoryAlloc))
	fmt.Printf("Memory Total: %s\n", stats.FormatBytes(stats.MemoryTotal))
	fmt.Printf("Memory System: %s\n", stats.FormatBytes(stats.MemorySys))
	fmt.Printf("GC Count: %d\n", stats.NumGC)
	fmt.Printf("GC Pause Total: %v\n", time.Duration(stats.GCPauseTotal))
	fmt.Printf("Last GC: %v\n", stats.LastGC)
	fmt.Println("========================\n")
}