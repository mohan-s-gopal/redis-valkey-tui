package ui

import (
	"fmt"
	"time"

	"redis-cli-dashboard/internal/logger"
	"redis-cli-dashboard/internal/utils"

	"github.com/dustin/go-humanize"
	"github.com/rivo/tview"
)

// createHeader creates the Redis-style header bar
func (a *App) createHeader() *tview.Flex {
	header := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetWrap(false)

	// Set the initial header content
	header.SetBorder(true)
	header.SetBorderPadding(0, 0, 1, 1)
	a.updateHeaderContent(header)

	// Create a flex container for the header
	headerFlex := tview.NewFlex().
		AddItem(header, 0, 1, false)

	// Update header content periodically
	stopChan := make(chan struct{})

	// Start the metrics collection routine
	go func() {
		ticker := time.NewTicker(2 * time.Second) // Update every 2 seconds
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Update header text directly (avoid QueueUpdateDraw hanging issues)
				if a.app != nil && header != nil {
					a.updateHeaderContent(header)
				}
			case <-stopChan:
				return
			}
		}
	}()

	// Store stop channel in app for cleanup
	a.metricsStopChan = stopChan

	return headerFlex
}

// formatHeaderText formats the header text based on current metrics
func (a *App) formatHeaderText() string {
	info, err := a.redis.Info()
	if err != nil {
		return "[red]Disconnected from Redis"
	}

	// Get key count using DBSIZE command (more accurate than parsing INFO)
	keyCount := 0
	if dbSize, err := a.redis.DBSize(); err == nil {
		keyCount = int(dbSize)
	}

	// Parse Redis version
	redisVersion, _ := info["redis_version"].(string)
	if redisVersion == "" {
		redisVersion = "unknown"
	}

	// Parse eviction policy
	evictionPolicy, _ := info["maxmemory_policy"].(string)
	if evictionPolicy == "" {
		evictionPolicy = "noeviction"
	}

	// Parse server state/role
	redisMode, _ := info["redis_mode"].(string)
	if redisMode == "" {
		redisMode = "standalone"
	}

	// Get role information for more detailed state
	redisRole, _ := info["role"].(string)
	if redisRole == "" {
		redisRole = "master"
	}

	// Combine mode and role for comprehensive state
	redisState := fmt.Sprintf("%s-%s", redisMode, redisRole)
	if redisMode == "standalone" {
		redisState = redisRole // Just show role for standalone
	}

	// Parse other metrics from info
	usedMemory, _ := info["used_memory"].(int64)
	connectedClients, _ := info["connected_clients"].(int64)
	uptimeSeconds, _ := info["uptime_in_seconds"].(int64)

	memory := humanize.Bytes(uint64(usedMemory))
	uptime := utils.FormatUptime(uptimeSeconds)

	return fmt.Sprintf(" redis-dashboard │ DB: db%d │ Keys: %d │ Version: %s │ State: %s │ Eviction: %s │ Memory: %s │ Clients: %d │ Uptime: %s │ [dim]1-6: Views │ ?: Help[white] ",
		a.config.Redis.DB,
		keyCount,
		redisVersion,
		redisState,
		evictionPolicy,
		memory,
		connectedClients,
		uptime)
}

// updateHeaderContent updates the header content with Redis metrics
func (a *App) updateHeaderContent(header *tview.TextView) {
	header.SetText(a.formatHeaderText())
}

// updateHeaderStatus updates the header status line with current metrics
func (a *App) updateHeaderStatus(status *tview.TextView) {
	text := fmt.Sprintf(
		"[white]ⓘ redis://%s:%d/%d  ∞ %dms  ↑%s/↓%s  ⚡ %d ops  ⚪ %d clients",
		a.config.Redis.Host,
		a.config.Redis.Port,
		a.config.Redis.DB,
		a.metrics.Latency,
		humanize.Bytes(uint64(a.metrics.MemoryUsed)),
		humanize.Bytes(uint64(a.metrics.MemoryPeak)),
		a.metrics.OpsPerSec,
		a.metrics.ConnectedClients,
	)
	status.SetText(text)
}

// updateContextLine updates the context line with current view info
func (a *App) updateContextLine(context *tview.TextView) {
	var viewContext string
	switch a.currentView {
	case KeysViewType:
		viewContext = fmt.Sprintf("[yellow]Context: Keys(%d) [white]🔍 Filter: %s",
			a.keysView.GetKeyCount(),
			a.keysView.GetFilter())
	case InfoViewType:
		viewContext = "[yellow]Context: Info"
	case MonitorViewType:
		viewContext = "[yellow]Context: Monitor"
	case CLIViewType:
		viewContext = "[yellow]Context: CLI"
	case ConfigViewType:
		viewContext = "[yellow]Context: Config"
	}

	viewActions := fmt.Sprintf("[white]1:[yellow]Keys [white]2:[yellow]Monitor [white]3:[yellow]Info [white]4:[yellow]CLI [white]5:[yellow]Config")
	keyActions := fmt.Sprintf("[white]<[yellow]a[white]>Add <[yellow]d[white]>Del <[yellow]r[white]>Refresh <[yellow]f[white]>Filter <[yellow]?[white]>Help")
	context.SetText(fmt.Sprintf("%s     %s     %s", viewContext, viewActions, keyActions))
}

// updateMetrics fetches the latest metrics from Redis
func (a *App) updateMetrics() {
	// Create a channel to handle timeout
	done := make(chan bool, 1)
	var info map[string]interface{}
	var err error

	// Run Redis INFO command with timeout
	go func() {
		info, err = a.redis.Info()
		done <- true
	}()

	// Wait for Redis operation with timeout
	select {
	case <-done:
		if err != nil {
			logger.Logger.Printf("Error fetching Redis info: %v", err)
			return
		}
	case <-time.After(500 * time.Millisecond):
		logger.Logger.Println("Timeout while fetching Redis metrics")
		return
	}

	// Update metrics from Redis INFO command
	if val, ok := info["latency_ms"].(int64); ok {
		a.metrics.Latency = val
	}
	if val, ok := info["used_memory"].(int64); ok {
		a.metrics.MemoryUsed = uint64(val)
	}
	if val, ok := info["used_memory_peak"].(int64); ok {
		a.metrics.MemoryPeak = uint64(val)
	}
	if val, ok := info["instantaneous_ops_per_sec"].(int64); ok {
		a.metrics.OpsPerSec = val
	}
	if val, ok := info["connected_clients"].(int64); ok {
		a.metrics.ConnectedClients = val
	}
}
