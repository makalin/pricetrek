package scheduler

import (
	"fmt"
	"runtime"
)

type Scheduler struct{}

func New() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) GenerateHourlySchedule() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return s.generateLaunchdHourly(), nil
	case "linux":
		return s.generateSystemdHourly(), nil
	case "windows":
		return s.generateWindowsHourly(), nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func (s *Scheduler) GenerateDailySchedule() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return s.generateLaunchdDaily(), nil
	case "linux":
		return s.generateSystemdDaily(), nil
	case "windows":
		return s.generateWindowsDaily(), nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func (s *Scheduler) generateLaunchdHourly() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.user.pricetrek</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/pricetrek</string>
        <string>track</string>
        <string>--once</string>
    </array>
    <key>StartInterval</key>
    <integer>3600</integer>
    <key>RunAtLoad</key>
    <true/>
    <key>StandardOutPath</key>
    <string>~/pricetrek/cron.log</string>
    <key>StandardErrorPath</key>
    <string>~/pricetrek/cron.log</string>
</dict>
</plist>`
}

func (s *Scheduler) generateLaunchdDaily() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.user.pricetrek</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/pricetrek</string>
        <string>track</string>
        <string>--once</string>
    </array>
    <key>StartCalendarInterval</key>
    <dict>
        <key>Hour</key>
        <integer>9</integer>
        <key>Minute</key>
        <integer>0</integer>
    </dict>
    <key>RunAtLoad</key>
    <true/>
    <key>StandardOutPath</key>
    <string>~/pricetrek/cron.log</string>
    <key>StandardErrorPath</key>
    <string>~/pricetrek/cron.log</string>
</dict>
</plist>`
}

func (s *Scheduler) generateSystemdHourly() string {
	return `[Unit]
Description=PriceTrek Price Tracker
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/pricetrek track --once
User=%i
WorkingDirectory=%h/pricetrek
StandardOutput=append:%h/pricetrek/cron.log
StandardError=append:%h/pricetrek/cron.log

[Install]
WantedBy=default.target

---
[Unit]
Description=PriceTrek Price Tracker Timer
Requires=pricetrek.service

[Timer]
OnCalendar=hourly
Persistent=true

[Install]
WantedBy=timers.target`
}

func (s *Scheduler) generateSystemdDaily() string {
	return `[Unit]
Description=PriceTrek Price Tracker
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/pricetrek track --once
User=%i
WorkingDirectory=%h/pricetrek
StandardOutput=append:%h/pricetrek/cron.log
StandardError=append:%h/pricetrek/cron.log

[Install]
WantedBy=default.target

---
[Unit]
Description=PriceTrek Price Tracker Timer
Requires=pricetrek.service

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target`
}

func (s *Scheduler) generateWindowsHourly() string {
	return `schtasks /Create /SC HOURLY /MO 1 /TN "PriceTrek" /TR "C:\\pricetrek\\pricetrek.exe track --once" /F`
}

func (s *Scheduler) generateWindowsDaily() string {
	return `schtasks /Create /SC DAILY /TN "PriceTrek" /TR "C:\\pricetrek\\pricetrek.exe track --once" /F`
}