# PriceTrek

> A comprehensive, fast terminal agent to **track product prices** on your watchlist and alert you on drops — **hourly or daily**. Works with web stores (selectors/APIs), saves history, plugs into your favorite notifiers, and provides advanced analytics and system monitoring.

![Go](https://img.shields.io/badge/Go-%3E=1.22-blue)
![License](https://img.shields.io/badge/License-MIT-green)
![Platform](https://img.shields.io/badge/OS-macOS%20%7C%20Linux%20%7C%20Windows-informational)
![Status](https://img.shields.io/badge/Status-Beta-green)
![Features](https://img.shields.io/badge/Features-15%2B-brightgreen)

---

## Highlights

* **🖥️ Rich CLI**: 15+ commands with comprehensive flag support
* **📊 Advanced Analytics**: Sparklines, statistics, price trends, moving averages
* **🔄 Smart Scheduling**: Cross-platform (cron/systemd/launchd/Task Scheduler)
* **🌐 Multi-Store Support**: CSS selectors, JSON APIs, custom provider scripts
* **⚡ Resilient Scraping**: Retries, backoff, jitter, HTTP caching, rate limiting
* **💾 Robust Storage**: SQLite with CSV/YAML import/export
* **🔔 Rich Notifications**: Email, Telegram, Slack, Ntfy with custom templates
* **📈 Data Visualization**: Terminal charts, sparklines, formatted reports
* **🛠️ System Monitoring**: Real-time stats, health checks, performance metrics
* **💾 Backup & Restore**: Compressed archives with automated cleanup
* **🔧 Developer Tools**: Makefile, Docker, comprehensive logging
* **📚 Extensible**: Clean interfaces, modular architecture, easy to extend

---

## Quick Start

```bash
# 1) Install (placeholder)
go install github.com/makalin/pricetrek@latest

# 2) Create a workspace
mkdir ~/pricetrek && cd ~/pricetrek
pricetrek init  # writes pricetrek.yaml and creates data/trek.db

# 3) Add a product
pricetrek add \
  --name "Samsung 990 Pro 2TB" \
  --url "https://www.hepsiburada.com/..." \
  --provider generic \
  --selector ".price .value" \
  --currency TRY \
  --target 4250

# 4) Run once (test)
pricetrek track --once --verbose

# 5) Set schedule (hourly example)
pricetrek schedule --hourly
```

---

## Configuration

`pricetrek.yaml` (auto-created by `init`)

```yaml
storage:
  driver: sqlite
  path: ./data/trek.db   # fallback: ./data/history.csv if sqlite not available

defaults:
  currency: TRY
  timezone: Europe/Istanbul
  user_agent: "PriceTrek/0.1 (+https://github.com/yourname/pricetrek)"
  retry:
    attempts: 3
    base_delay_ms: 800
    max_delay_ms: 7000
  http_timeout_sec: 20
  cache_ttl_min: 30
  headless:
    enabled: false         # set true for JS-heavy pages (uses Playwright)
    wait_until: "networkidle"

notifications:
  # enable any you like (leave secrets in env)
  email:
    enabled: false
    from: "alerts@pricetrek.local"
    to: ["me@example.com"]
  telegram:
    enabled: false
    chat_id: "123456789"
  slack:
    enabled: false
    webhook: ""            # or set PRICETREK_SLACK_WEBHOOK
  ntfy:
    enabled: false
    topic: "pricetrek"

rules:
  # global fallbacks used if item has no rule
  percent_drop: 8          # alert if price falls >= 8%
  target_price: null       # optional global target (overridden per item)

items:
  - id: "990pro-2tb"
    name: "Samsung 990 Pro 2TB"
    url: "https://www.hepsiburada.com/..."
    provider: generic
    selector: ".product-price .value"   # CSS selector (textContent parsed as number)
    currency: TRY
    target_price: 4250
    percent_drop: 10
    schedule: "hourly"                  # hourly | daily | cron("*/15 * * * *")
  - id: "ps5-slim"
    name: "PS5 Slim"
    url: "https://www.trendyol.com/..."
    provider: generic
    selector: "span.prc-dsc"
    currency: TRY
    schedule: "daily"
```

> **Secrets via ENV**
> `PRICETREK_EMAIL_SMTP`, `PRICETREK_EMAIL_USER`, `PRICETREK_EMAIL_PASS`,
> `PRICETREK_TELEGRAM_TOKEN`, `PRICETREK_SLACK_WEBHOOK`, `PRICETREK_NTFY_URL`, etc.

---

## CLI Commands

### Core Commands
```text
pricetrek init                       # Initialize workspace and configuration
pricetrek add --name --url ...       # Add product with full flag support
pricetrek rm <id>                    # Remove item with confirmation
pricetrek ls [--json] [--verbose]    # List watchlist with detailed info
pricetrek show <id> [--spark]        # Price history with sparklines & stats
pricetrek track [--once|--loop]      # Run tracking with caching options
pricetrek alert --dry-run            # Check and send price alerts
```

### Data Management
```text
pricetrek export --csv file [--items|--prices]  # Export data to CSV
pricetrek import --csv file [--yaml file]       # Import from CSV/YAML
pricetrek backup [--output file] [--dir dir]    # Create compressed backup
pricetrek restore --file backup [--target dir]  # Restore from backup
```

### System & Monitoring
```text
pricetrek doctor                     # Comprehensive health check
pricetrek schedule --hourly|--daily  # Generate OS-specific schedules
pricetrek monitor [--once] [--interval] # System performance monitoring
pricetrek help                       # Show detailed help
```

### Examples

* **Add a product with all options**:
```bash
pricetrek add \
  --name "Samsung 990 Pro 2TB" \
  --url "https://www.hepsiburada.com/..." \
  --provider generic \
  --selector ".price .value" \
  --currency TRY \
  --target 4250 \
  --percent 10 \
  --schedule hourly
```

* **Track with caching and specific item**:
```bash
pricetrek track --once --respect-cache --id 990pro-2tb
```

* **Show detailed price history with sparkline**:
```bash
pricetrek show 990pro-2tb --spark --limit 50
# Price Trend: ▁▂▃▄▅▆█▇▆▅▄▃▂▁ (last 50 samples)
```

* **Export data and create backup**:
```bash
pricetrek export --csv prices.csv --prices
pricetrek backup --output backup-$(date +%Y%m%d).tar.gz
```

* **Monitor system performance**:
```bash
pricetrek monitor --once
pricetrek doctor
```

* **Generate scheduling for your OS**:
```bash
pricetrek schedule --hourly > ~/Library/LaunchAgents/com.user.pricetrek.plist
```

---

## Providers

PriceTrek supports two paths:

1. **Generic (selector)** — good for static pages

```yaml
provider: generic
selector: "span.price, .amount"
attr: "text"           # or "content", "data-price"
regex: "([0-9][0-9\\.,]+)"    # optional cleanup
```

2. **Custom Providers** — for JSON APIs or complex sites
   Implement a tiny binary/script that prints a price JSON to stdout:

```json
{"price": 4199.00, "currency": "TRY", "in_stock": true, "extra": {"seller":"ACME"}}
```

Then reference it:

```yaml
provider: exec
command: "./providers/trendyol.sh {{url}}"
```

> Bundled templates: **Amazon**, **eBay**, **AliExpress**, **Hepsiburada**, **Trendyol**, **Newegg** (via selectors or light API wrappers where available).

---

## Storage Model

* **SQLite** table `prices(item_id TEXT, ts DATETIME, price REAL, currency TEXT, meta JSON)`
* Rolling **stats**: min / max / 7-day Δ / 30-day Δ
* Auto **FX normalize** (optional): set `fx.base = USD|EUR|TRY` and feed rates via `fx.rates_url` or manual table.

Example schema (for reference):

```sql
CREATE TABLE IF NOT EXISTS prices (
  item_id TEXT NOT NULL,
  ts      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  price   REAL NOT NULL,
  currency TEXT NOT NULL,
  meta    TEXT
);
CREATE INDEX IF NOT EXISTS idx_prices_item_ts ON prices(item_id, ts DESC);
```

---

## Scheduling

### macOS (launchd)

```bash
pricetrek schedule --hourly > ~/Library/LaunchAgents/com.user.pricetrek.plist
launchctl load ~/Library/LaunchAgents/com.user.pricetrek.plist
```

### Linux (systemd timer)

```bash
pricetrek schedule --hourly --systemd
# creates: ~/.config/systemd/user/pricetrek.service & pricetrek.timer
systemctl --user enable --now pricetrek.timer
```

### Linux/Unix (cron)

```bash
crontab -e
# run at minute 7 every hour
7 * * * * /usr/local/bin/pricetrek track --once >> ~/pricetrek/cron.log 2>&1
```

### Windows (Task Scheduler)

```powershell
# hourly
schtasks /Create /SC HOURLY /MO 1 /TN "PriceTrek" /TR "C:\pricetrek\pricetrek.exe track --once"
```

---

## Alerts

Rules are evaluated on each new sample:

* `target_price` met or beaten
* `percent_drop` relative to last N samples (default N=3)
* `in_stock` flipped from false→true (optional)

Templates:

```text
[PriceTrek] PS5 Slim ↓ 6% to 21.999,00 TRY (Trendyol)
Prev: 23.399,00 TRY (−1.400,00) • 7d Δ −8.4%
https://www.trendyol.com/...
```

Enable notifiers in `pricetrek.yaml` and/or via ENV.
Examples:

```bash
export PRICETREK_TELEGRAM_TOKEN=123:ABC
export PRICETREK_SLACK_WEBHOOK="https://hooks.slack.com/..."
pricetrek alert --test "Hello from PriceTrek"
```

---

## Resilience & Ethics

* Polite: randomized delays, capped concurrency, `If-Modified-Since`/ETag
* Respect store terms; prefer official APIs when available
* Headless only when necessary; exponential backoff on errors
* Local cache with TTL to avoid hammering sites

---

## Import / Export

```bash
# import a CSV watchlist
pricetrek import --csv items.csv

# export full history
pricetrek export --csv history.csv
```

`items.csv` columns: `id,name,url,provider,selector,currency,target_price,percent_drop,schedule`

---

## Troubleshooting

```bash
pricetrek doctor
# checks: DB, network, DNS, headless binary, selectors, notifiers, fx source
```

Common fixes:

* JS-heavy page → set `headless.enabled: true`
* Wrong number parsing → add `regex` cleanup
* Currency symbol issue → set `currency` explicitly
* No alerts → check `rules`, thresholds, and notifier env vars

---

## What's New in v0.2.0

### 🆕 Major Features Added
- **System Monitoring**: Real-time performance metrics and health checks
- **Backup & Restore**: Compressed data archives with automated cleanup
- **Advanced Analytics**: Sparklines, statistics, and price trend analysis
- **Enhanced CLI**: 15+ commands with comprehensive flag support
- **Data Management**: CSV/YAML import/export with validation
- **Rich Notifications**: Email, Telegram, Slack, Ntfy with templates
- **Developer Tools**: Makefile, Docker, structured logging

### 📊 Statistics
- **21 Go files** with 3,388+ lines of code
- **8 internal packages** with specialized functionality
- **4 notification providers** fully implemented
- **3 scheduling systems** (macOS, Linux, Windows)
- **2 data formats** (CSV, YAML) supported

## Roadmap

* [ ] Price charts (`pricetrek graph <id>` with PNG/terminal output)
* [ ] ML-based “deal score” (seasonality + competitor diff)
* [ ] Browser extension to “Send to PriceTrek”
* [ ] Encrypted cloud sync (S3/Gist) for watchlist only
* [ ] Multi-currency arbitrage view

---

## Advanced Features

### System Monitoring
- **Real-time statistics**: Memory usage, CPU, goroutines, GC metrics
- **Health diagnostics**: Database, network, provider validation
- **Performance tracking**: Allocation patterns, garbage collection stats
- **Continuous monitoring**: Configurable intervals with live updates

### Data Management
- **CSV Import/Export**: Universal data exchange format
- **YAML Configuration**: Human-readable configuration files
- **Compressed Backups**: Tar.gz archives with metadata
- **Data Validation**: Input sanitization and error handling

### Analytics & Visualization
- **Sparklines**: ASCII price trend visualization
- **Price Statistics**: Min, max, average, median calculations
- **Moving Averages**: Trend analysis and smoothing
- **Currency Formatting**: Multi-currency support with symbols

### Developer Tools
- **Makefile**: Complete development workflow
- **Docker Support**: Containerized deployment
- **Comprehensive Logging**: Structured logging with levels
- **Health Checks**: Automated system validation

## Development

```bash
git clone https://github.com/makalin/pricetrek
cd pricetrek
make dev      # builds pricetrek (Go 1.22+), runs linters & tests
make build    # build for current platform
make build-all # build for all platforms
make test     # run tests with coverage
make docker-build # build Docker image
```

### Project Structure
```
pricetrek/
├── internal/
│   ├── cli/          # Command-line interface
│   ├── config/       # Configuration management
│   ├── csv/          # CSV import/export
│   ├── logger/       # Structured logging
│   ├── notifications/# Notification providers
│   ├── providers/    # Price provider interfaces
│   ├── scheduler/    # Cross-platform scheduling
│   ├── storage/      # Data storage layer
│   ├── tools/        # Backup, monitoring utilities
│   ├── tracker/      # Price tracking logic
│   └── utils/        # Sparklines, formatting
├── examples/         # Sample configurations
├── scripts/          # Installation and testing
└── build/           # Build artifacts
```

### Provider Interface
```go
type Provider interface {
    Fetch(ctx context.Context, item Item) (PriceSample, error)
}

type PriceSample struct {
    Price    float64                `json:"price"`
    Currency string                 `json:"currency"`
    InStock  bool                   `json:"in_stock"`
    Meta     map[string]interface{} `json:"meta,omitempty"`
}
```

---

## License

MIT © Mehmet T. AKALIN. Contributions welcome.
