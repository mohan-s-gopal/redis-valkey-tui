package ui

import (
	"fmt"
	"github.com/mohan-s-gopal/redis-valkey-tui/internal/redis"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// InfoView represents the server info view
type InfoView struct {
	redis       *redis.Client
	flex        *tview.Flex
	infoTable   *tview.Table
	metricsText *tview.TextView
}

// NewInfoView creates a new info view
func NewInfoView(redisClient *redis.Client) *InfoView {
	view := &InfoView{
		redis: redisClient,
	}

	view.setupUI()
	view.loadInfo()

	return view
}

// setupUI initializes the UI components
func (v *InfoView) setupUI() {
	// Create server info table (left side, like redis-cli --stat)
	v.infoTable = tview.NewTable()
	v.infoTable.SetBorder(true).
		SetTitle("Redis Server Information").
		SetTitleAlign(tview.AlignLeft)
	v.infoTable.SetSelectable(false, false)

	// Create metrics display (right side)
	v.metricsText = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)
	v.metricsText.SetBorder(true).
		SetTitle("Performance Metrics").
		SetTitleAlign(tview.AlignLeft)

	// Create horizontal layout
	v.flex = tview.NewFlex().SetDirection(tview.FlexColumn)
	v.flex.AddItem(v.infoTable, 0, 1, true)
	v.flex.AddItem(v.metricsText, 0, 1, false)

	// Set up input capture for the main flex
	v.flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle view-specific keys
		switch event.Rune() {
		case 'r', 'R':
			// Refresh info
			v.loadInfo()
			return nil
		}

		// Let all other keys pass through to global handler
		return event
	})
}

// GetComponent returns the main component
func (v *InfoView) GetComponent() tview.Primitive {
	return v.flex
}

// loadInfo loads server information
func (v *InfoView) loadInfo() {
	info, err := v.redis.GetInfo()
	if err != nil {
		// Show error in table
		v.infoTable.Clear()
		v.infoTable.SetCell(0, 0, tview.NewTableCell("Error").SetTextColor(tcell.ColorRed))
		v.infoTable.SetCell(0, 1, tview.NewTableCell(fmt.Sprintf("Failed to load: %v", err)).SetTextColor(tcell.ColorRed))
		v.metricsText.SetText(fmt.Sprintf("[red]Error loading server info: %s", err))
		return
	}

	v.loadServerInfoTable(info)
	v.loadMetricsDisplay(info)
}

// loadServerInfoTable populates the server info table like redis-cli --stat
func (v *InfoView) loadServerInfoTable(info map[string]string) {
	v.infoTable.Clear()

	// Key server information in the order shown in your screenshot
	serverInfo := [][]string{
		{"redis_version", getInfoValueStr(info, "redis_version", "unknown")},
		{"process_id", getInfoValueStr(info, "process_id", "unknown")},
		{"uptime_in_seconds", getInfoValueStr(info, "uptime_in_seconds", "0")},
		{"uptime_in_days", getInfoValueStr(info, "uptime_in_days", "0")},
		{"gcc_version", getInfoValueStr(info, "gcc_version", "unknown")},
		{"role", getInfoValueStr(info, "role", "unknown")},
		{"connected_slaves", getInfoValueStr(info, "connected_slaves", "0")},
		{"aof_enabled", getInfoValueStr(info, "aof_enabled", "0")},
		{"vm_enabled", getInfoValueStr(info, "vm_enabled", "0")},
		{"tcp_port", getInfoValueStr(info, "tcp_port", "6379")},
		{"config_file", getInfoValueStr(info, "config_file", "")},
		{"os", getInfoValueStr(info, "os", "unknown")},
		{"arch_bits", getInfoValueStr(info, "arch_bits", "unknown")},
		{"multiplexing_api", getInfoValueStr(info, "multiplexing_api", "unknown")},
		{"process_id", getInfoValueStr(info, "process_id", "unknown")},
		{"used_memory", getInfoValueStr(info, "used_memory", "0")},
		{"used_memory_human", getInfoValueStr(info, "used_memory_human", "0B")},
		{"used_memory_peak", getInfoValueStr(info, "used_memory_peak", "0")},
		{"used_memory_peak_human", getInfoValueStr(info, "used_memory_peak_human", "0B")},
		{"mem_fragmentation_ratio", getInfoValueStr(info, "mem_fragmentation_ratio", "0")},
		{"changes_since_last_save", getInfoValueStr(info, "changes_since_last_save", "0")},
		{"last_save_time", getInfoValueStr(info, "last_save_time", "0")},
		{"total_connections_received", getInfoValueStr(info, "total_connections_received", "0")},
		{"total_commands_processed", getInfoValueStr(info, "total_commands_processed", "0")},
		{"expired_keys", getInfoValueStr(info, "expired_keys", "0")},
		{"evicted_keys", getInfoValueStr(info, "evicted_keys", "0")},
		{"keyspace_hits", getInfoValueStr(info, "keyspace_hits", "0")},
		{"keyspace_misses", getInfoValueStr(info, "keyspace_misses", "0")},
		{"pubsub_channels", getInfoValueStr(info, "pubsub_channels", "0")},
		{"pubsub_patterns", getInfoValueStr(info, "pubsub_patterns", "0")},
	}

	for i, info := range serverInfo {
		// Left column: property name
		v.infoTable.SetCell(i, 0, tview.NewTableCell(info[0]).SetTextColor(tcell.ColorGreen))
		// Right column: value
		v.infoTable.SetCell(i, 1, tview.NewTableCell(info[1]).SetTextColor(tcell.ColorWhite))
	}
}

// loadMetricsDisplay loads the performance metrics like in redis-cli --stat
func (v *InfoView) loadMetricsDisplay(info map[string]string) {
	// Get current metrics
	metrics, err := v.redis.GetMetrics()
	if err != nil {
		v.metricsText.SetText(fmt.Sprintf("[red]Error loading metrics: %s", err))
		return
	}

	// Format like redis-cli --stat with columns
	metricsText := `[yellow]Redis Performance Statistics[white]

[cyan]━━━ Current Stats ━━━[white]
[green]Connected Clients:[white] %d
[green]Used Memory:[white] %s
[green]Used Memory RSS:[white] %s
[green]Total Commands:[white] %d
[green]Ops/sec:[white] %d
[green]Keyspace Hits:[white] %d
[green]Keyspace Misses:[white] %d

[cyan]━━━ Hit Rate ━━━[white]
[green]Hit Rate:[white] %.2f%%

[cyan]━━━ Persistence ━━━[white]
[green]RDB Last Save:[white] %s
[green]AOF Enabled:[white] %s
[green]Changes Since Save:[white] %s

[cyan]━━━ Connections ━━━[white]
[green]Total Connections:[white] %s
[green]Rejected Connections:[white] %s

[cyan]━━━ Key Events ━━━[white]
[green]Expired Keys:[white] %s
[green]Evicted Keys:[white] %s

[cyan]━━━ Replication ━━━[white]
[green]Role:[white] %s
[green]Connected Slaves:[white] %s
`

	// Calculate hit rate
	hitRate := float64(0)
	if metrics.KeyspaceHits+metrics.KeyspaceMisses > 0 {
		hitRate = float64(metrics.KeyspaceHits) / float64(metrics.KeyspaceHits+metrics.KeyspaceMisses) * 100
	}

	// Format uptime (not currently used but useful for future enhancements)
	_ = getInfoValueStr(info, "uptime_in_seconds", "0")

	formattedText := fmt.Sprintf(metricsText,
		metrics.ConnectedClients,
		getInfoValueStr(info, "used_memory_human", "0B"),
		getInfoValueStr(info, "used_memory_rss_human", "0B"),
		metrics.TotalCommandsProcessed,
		metrics.InstantaneousOpsPerSec,
		metrics.KeyspaceHits,
		metrics.KeyspaceMisses,
		hitRate,
		getInfoValueStr(info, "last_save_time", "0"),
		getInfoValueStr(info, "aof_enabled", "0"),
		getInfoValueStr(info, "changes_since_last_save", "0"),
		getInfoValueStr(info, "total_connections_received", "0"),
		getInfoValueStr(info, "rejected_connections", "0"),
		getInfoValueStr(info, "expired_keys", "0"),
		getInfoValueStr(info, "evicted_keys", "0"),
		getInfoValueStr(info, "role", "master"),
		getInfoValueStr(info, "connected_slaves", "0"),
	)

	v.metricsText.SetText(formattedText)
}

// getInfoValueStr safely extracts a string value from info map
func getInfoValueStr(info map[string]string, key, defaultValue string) string {
	if val, ok := info[key]; ok {
		return val
	}
	return defaultValue
}

// Refresh refreshes the info view
func (v *InfoView) Refresh() {
	v.loadInfo()
}
