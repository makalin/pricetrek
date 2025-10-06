package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/makalin/pricetrek/internal/cli"
	"github.com/makalin/pricetrek/internal/config"
	"github.com/makalin/pricetrek/internal/logger"
)

const version = "0.1.0"

func main() {
	var (
		configPath = flag.String("config", "pricetrek.yaml", "Path to configuration file")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
		versionFlag = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *versionFlag {
		fmt.Printf("PriceTrek v%s\n", version)
		os.Exit(0)
	}

	// Initialize logger
	log := logger.New(*verbose)

	// Parse command line arguments
	args := flag.Args()
	if len(args) == 0 {
		// Show help without requiring config
		cli := cli.New(nil, log)
		cli.Help()
		os.Exit(1)
	}

	// Handle init command without requiring config
	if args[0] == "init" {
		cli := cli.New(nil, log)
		ctx := context.Background()
		if err := cli.Execute(ctx, args); err != nil {
			log.Error("Command failed", "error", err)
			os.Exit(1)
		}
		return
	}

	// Load configuration for other commands
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	// Create CLI instance
	cli := cli.New(cfg, log)

	// Execute command
	ctx := context.Background()
	if err := cli.Execute(ctx, args); err != nil {
		log.Error("Command failed", "error", err)
		os.Exit(1)
	}
}