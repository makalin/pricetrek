# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project structure with Go modules
- Comprehensive `.gitignore` file for Go development
- CLI framework with command structure
- Configuration management with YAML support
- SQLite storage backend with fallback support
- Generic web scraping provider
- HTTP client for web requests
- Notification system framework (email, telegram, slack, ntfy)
- Scheduling system (cron, systemd, launchd, Windows Task Scheduler)
- Docker support with multi-stage builds
- Makefile for development workflow
- Installation script for Unix-like systems
- Provider testing script
- Example configuration file

### Technical Details
- Go 1.22+ support
- SQLite database with WAL mode
- Structured logging with slog
- Context-aware operations
- Modular architecture with clean interfaces
- Cross-platform support (macOS, Linux, Windows)
- Docker containerization ready

## [0.1.0] - 2025-10-06

### Added
- Initial release
- Basic CLI commands (init, help)
- Configuration file generation
- Database initialization
- Project scaffolding