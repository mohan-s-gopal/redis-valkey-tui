package ui

import (
	"fmt"
	"time"

	"redis-cli-dashboard/internal/utils"
	"redis-cli-dashboard/internal/logger"

	"github.com/dustin/go-humanize"
	"github.com/rivo/tview"
)


// createHeader creates the Redis-style header bar
func (a *App) createHeader() *tview.Flex {
	header := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	// Set the initial header content
	header.SetBorder(true)
	a.updateHeaderContent(header)

	// Create a flex container for the header
	headerFlex := tview.NewFlex().
		AddItem(header, 0, 1, false)

	// Update metrics periodically
	// stopChan := make(chan struct{})
	// updateChan := make(chan struct{}, 1) // Buffer of 1 to prevent blocking

	// Start the metrics collection routine
	// go func() {
	// 	ticker := time.NewTicker(time.Second)
	// 	defer ticker.Stop()
	// 	logger.Logger.Println("Metrics update routine started")

	// 	metricsUpdateChan := make(chan struct{}, 1) // Channel for metrics updates
	// 	go func() {
	// 		for range metricsUpdateChan {
	// 			a.updateMetrics()
	// 			select {
	// 			case updateChan <- struct{}{}:
	// 			default:
	// 			}
	// 		}
	// 	}()

	// 	for {
	// 		select {
	// 		case <-ticker.C:
	// 			// Non-blocking send to metrics update channel
	// 			select {
	// 			case metricsUpdateChan <- struct{}{}:
	// 			default:
	// 				// Skip this update if previous one is still processing
	// 				logger.Logger.Println("Skipping metrics update - previous update still in progress")
	// 			}
	// 		case <-stopChan:
	// 			close(metricsUpdateChan)
	// 			logger.Logger.Println("Metrics update routine stopped")
	// 			return
	// 		}
	// 	}
	// }()

	// Start the UI update routine
	// go func() {
	// 	logger.Logger.Println("UI update routine started")
	// 	for {
	// 		select {
	// 		case <-updateChan:
	// 			header.SetText(a.formatHeaderText())
	// 		case <-stopChan:
	// 			logger.Logger.Println("UI update routine stopped")
	// 			return
	// 		}
	// 	}
	// }()

	// Store stop channel in app for cleanup
	// a.metricsStopChan = stopChan

	return headerFlex
}

// formatHeaderText formats the header text based on current metrics
func (a *App) formatHeaderText() string {
	info, err := a.redis.Info()
	if err != nil {
		return "[red]Disconnected from Redis"
	}

	// Get key count
	keyCount := 0
	if dbSize, ok := info["db0_keys"].(int64); ok {
		keyCount = int(dbSize)
	}

	// Parse metrics from info
	usedMemory, _ := info["used_memory"].(int64)
	connectedClients, _ := info["connected_clients"].(int64)
	uptimeSeconds, _ := info["uptime_in_seconds"].(int64)
	
	memory := humanize.Bytes(uint64(usedMemory))
	uptime := utils.FormatUptime(uptimeSeconds)

	return fmt.Sprintf(" redis-cli-dashboard │ DB: db%d │ Keys: %d │ Memory: %s │ Clients: %d │ Uptime: %s ",
		a.config.Redis.DB,
		keyCount,
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


