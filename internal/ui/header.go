package ui

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rivo/tview"
)


// createHeader creates the k9s-like header bar
func (a *App) createHeader() *tview.Flex {
	header := tview.NewFlex().SetDirection(tview.FlexRow)

	// Create a box for the header
	headerBox := tview.NewFlex().
		SetDirection(tview.FlexRow)
	headerBox.SetBorder(true).
		SetTitle("Redis Dashboard")

	// Top status line with Redis info and metrics
	statusLine := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	// Context line with current view info and actions
	contextLine := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)

	a.updateHeaderStatus(statusLine)
	a.updateContextLine(contextLine)

	// Add status and context lines to header box
	headerBox.AddItem(statusLine, 1, 0, false).
		AddItem(contextLine, 1, 0, false)

	// Add header box to main header flex
	header.AddItem(headerBox, 3, 0, false)

	// Update metrics periodically
	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			a.updateMetrics()
			a.app.QueueUpdateDraw(func() {
				a.updateHeaderStatus(statusLine)
				a.updateContextLine(contextLine)
			})
		}
	}()

	return header
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
	info, err := a.redis.Info()
	if err != nil {
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


