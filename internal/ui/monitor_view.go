package ui

import (
	"fmt"
	"github.com/mohan-s-gopal/redis-valkey-tui/internal/logger"
	"github.com/mohan-s-gopal/redis-valkey-tui/internal/redis"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// MonitorView represents the monitoring view
type MonitorView struct {
	redis     *redis.Client
	
	// UI Components
	flex          *tview.Flex
	statsTable    *tview.Table
	commandTable  *tview.Table
	clientTable   *tview.Table
	infoText      *tview.TextView

	// Monitoring state
	monitoring    bool
	ticker        *time.Ticker
	stopChan      chan bool
	refreshRate   time.Duration
	refreshIndex  int // Index for cycling through refresh rates
}

// NewMonitorView creates a new monitor view
func NewMonitorView(redisClient *redis.Client) *MonitorView {
	logger.Logger.Println("Initializing MonitorView...")
	view := &MonitorView{
		redis:        redisClient,
		stopChan:     make(chan bool),
		refreshRate:  2 * time.Second, // Default 2 seconds like top
		refreshIndex: 1,               // Start with 2 seconds
	}

	view.setupUI()
	view.loadData()
	
	// Start monitoring automatically like top/htop
	view.startMonitoring()
	
	logger.Logger.Println("MonitorView initialized")

	return view
}

// setupUI initializes the UI components
func (v *MonitorView) setupUI() {
	// Create command statistics table
	v.commandTable = tview.NewTable()
	v.commandTable.SetBorder(true).
		SetTitle("Command Statistics").
		SetTitleAlign(tview.AlignLeft)
	v.commandTable.SetSelectable(true, false)
	
	// Set command table headers
	headers := []string{"Command", "Number of calls", "Total Duration ↓", "Duration per call", "RejectedCalls", "FailedCalls", "CallsMaster"}
	for i, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)
		v.commandTable.SetCell(0, i, cell)
	}

	// Create client connections table
	v.clientTable = tview.NewTable()
	v.clientTable.SetBorder(true).
		SetTitle("Client connections").
		SetTitleAlign(tview.AlignLeft)
	v.clientTable.SetSelectable(true, false)
	
	// Set client table headers
	clientHeaders := []string{"Client", "Total duration", "Idle time ↓", "Last command"}
	for i, header := range clientHeaders {
		cell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)
		v.clientTable.SetCell(0, i, cell)
	}

	// Create server statistics table
	v.statsTable = tview.NewTable()
	v.statsTable.SetBorder(true).
		SetTitle("Server Statistics").
		SetTitleAlign(tview.AlignLeft)
	v.statsTable.SetSelectable(false, false)

	// Create info text view for additional details
	v.infoText = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)
	v.infoText.SetBorder(true).
		SetTitle("System Information").
		SetTitleAlign(tview.AlignLeft)

	// Create main layout with vertical arrangement
	v.flex = tview.NewFlex().SetDirection(tview.FlexRow)
	
	// Top section: Command table
	v.flex.AddItem(v.commandTable, 0, 2, true)
	
	// Middle section: Client connections table
	v.flex.AddItem(v.clientTable, 0, 2, false)
	
	// Bottom section: Server stats and system info side by side
	bottomFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	bottomFlex.AddItem(v.statsTable, 0, 1, false)
	bottomFlex.AddItem(v.infoText, 0, 1, false)
	
	v.flex.AddItem(bottomFlex, 0, 1, false)

	// Set input capture for the main flex
	v.flex.SetInputCapture(v.handleInput)
}

// GetComponent returns the main component
func (v *MonitorView) GetComponent() tview.Primitive {
	return v.flex
}

// handleInput handles input for the monitor view
func (v *MonitorView) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 's', 'S':
		v.toggleMonitoring()
		return nil
	case 'c', 'C':
		v.clearScreen()
		return nil
	case 'r', 'R':
		v.Refresh()
		return nil
	case 'd', 'D':
		// Change refresh delay like in htop
		v.cycleRefreshRate()
		return nil
	}

	// Let all other keys pass through to global handler (including 1-6, ?, etc.)
	return event
}

// toggleMonitoring starts or stops monitoring
func (v *MonitorView) toggleMonitoring() {
	if v.monitoring {
		v.stopMonitoring()
	} else {
		v.startMonitoring()
	}
}

// startMonitoring starts real-time monitoring
func (v *MonitorView) startMonitoring() {
	if v.monitoring {
		return
	}

	v.monitoring = true
	v.ticker = time.NewTicker(v.refreshRate)

	go func() {
		for {
			select {
			case <-v.ticker.C:
				v.loadData()
			case <-v.stopChan:
				return
			}
		}
	}()

	v.updateTitle()
}

// stopMonitoring stops real-time monitoring
func (v *MonitorView) stopMonitoring() {
	if !v.monitoring {
		return
	}

	v.monitoring = false
	if v.ticker != nil {
		v.ticker.Stop()
	}
	v.stopChan <- true

	v.updateTitle()
}

// cycleRefreshRate cycles through different refresh rates like htop
func (v *MonitorView) cycleRefreshRate() {
	refreshRates := []time.Duration{
		1 * time.Second,  // Fast
		2 * time.Second,  // Normal
		5 * time.Second,  // Slow
		10 * time.Second, // Very slow
	}
	
	v.refreshIndex = (v.refreshIndex + 1) % len(refreshRates)
	v.refreshRate = refreshRates[v.refreshIndex]
	
	// Restart monitoring with new rate if currently monitoring
	if v.monitoring {
		v.stopMonitoring()
		v.startMonitoring()
	}
	
	// Update title to show current refresh rate
	v.updateTitleWithRefreshRate()
}

// updateTitleWithRefreshRate updates titles to show refresh rate
func (v *MonitorView) updateTitleWithRefreshRate() {
	rateStr := fmt.Sprintf("%.0fs", v.refreshRate.Seconds())
	
	if v.monitoring {
		v.commandTable.SetTitle(fmt.Sprintf("Command Statistics [ACTIVE - %s]", rateStr))
		v.clientTable.SetTitle(fmt.Sprintf("Client connections [ACTIVE - %s]", rateStr))
		v.statsTable.SetTitle(fmt.Sprintf("Server Statistics [ACTIVE - %s]", rateStr))
		v.infoText.SetTitle(fmt.Sprintf("System Information [ACTIVE - %s]", rateStr))
	} else {
		v.commandTable.SetTitle(fmt.Sprintf("Command Statistics [STOPPED - %s]", rateStr))
		v.clientTable.SetTitle(fmt.Sprintf("Client connections [STOPPED - %s]", rateStr))
		v.statsTable.SetTitle(fmt.Sprintf("Server Statistics [STOPPED - %s]", rateStr))
		v.infoText.SetTitle(fmt.Sprintf("System Information [STOPPED - %s]", rateStr))
	}
}

// updateTitle updates the title based on monitoring state
func (v *MonitorView) updateTitle() {
	v.updateTitleWithRefreshRate()
}

// clearScreen clears all tables and text
func (v *MonitorView) clearScreen() {
	// Clear command table (keep headers)
	for row := v.commandTable.GetRowCount() - 1; row > 0; row-- {
		v.commandTable.RemoveRow(row)
	}
	
	// Clear client table (keep headers)
	for row := v.clientTable.GetRowCount() - 1; row > 0; row-- {
		v.clientTable.RemoveRow(row)
	}
	
	// Clear stats table
	v.statsTable.Clear()
	
	// Clear info text
	v.infoText.SetText("")
}

// loadData loads and displays all monitoring data
func (v *MonitorView) loadData() {
	v.loadCommandStats()
	v.loadClientConnections()
	v.loadServerStats()
	v.loadSystemInfo()
}

// loadCommandStats loads command statistics into the table
func (v *MonitorView) loadCommandStats() {
	stats, err := v.redis.GetCommandStats()
	if err != nil {
		// Show error in first data row
		v.commandTable.SetCell(1, 0, tview.NewTableCell("ERROR").SetTextColor(tcell.ColorRed))
		v.commandTable.SetCell(1, 1, tview.NewTableCell(fmt.Sprintf("Failed to load: %v", err)).SetTextColor(tcell.ColorRed))
		return
	}

	// Clear existing data rows (keep header)
	for row := v.commandTable.GetRowCount() - 1; row > 0; row-- {
		v.commandTable.RemoveRow(row)
	}

	// Sort stats by total duration (descending)
	// Simple bubble sort for small datasets
	for i := 0; i < len(stats)-1; i++ {
		for j := 0; j < len(stats)-i-1; j++ {
			if stats[j].TotalDuration < stats[j+1].TotalDuration {
				stats[j], stats[j+1] = stats[j+1], stats[j]
			}
		}
	}

	// Add data rows
	for i, stat := range stats {
		row := i + 1
		
		// Command name
		v.commandTable.SetCell(row, 0, tview.NewTableCell(stat.Command))
		
		// Number of calls
		v.commandTable.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%.0f", float64(stat.Calls))))
		
		// Total duration
		v.commandTable.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%.1f ms", stat.TotalDuration)))
		
		// Duration per call
		v.commandTable.SetCell(row, 3, tview.NewTableCell(fmt.Sprintf("%.1f ms", stat.DurationPerCall)))
		
		// Rejected calls
		v.commandTable.SetCell(row, 4, tview.NewTableCell(fmt.Sprintf("%d", stat.RejectedCalls)))
		
		// Failed calls
		v.commandTable.SetCell(row, 5, tview.NewTableCell(fmt.Sprintf("%d", stat.FailedCalls)))
		
		// CallsMaster (not available in standard Redis, show 0)
		v.commandTable.SetCell(row, 6, tview.NewTableCell("0"))
	}
}

// loadClientConnections loads client connection information into the table
func (v *MonitorView) loadClientConnections() {
	clients, err := v.redis.GetClientList()
	if err != nil {
		// Show error in first data row
		v.clientTable.SetCell(1, 0, tview.NewTableCell("ERROR").SetTextColor(tcell.ColorRed))
		v.clientTable.SetCell(1, 1, tview.NewTableCell(fmt.Sprintf("Failed to load: %v", err)).SetTextColor(tcell.ColorRed))
		return
	}

	// Clear existing data rows (keep header)
	for row := v.clientTable.GetRowCount() - 1; row > 0; row-- {
		v.clientTable.RemoveRow(row)
	}

	// Sort clients by idle time (descending) 
	for i := 0; i < len(clients)-1; i++ {
		for j := 0; j < len(clients)-i-1; j++ {
			if clients[j].Idle < clients[j+1].Idle {
				clients[j], clients[j+1] = clients[j+1], clients[j]
			}
		}
	}

	// Add data rows
	for i, client := range clients {
		row := i + 1
		
		// Client address (IP:Port format)
		clientAddr := client.Address
		if client.Name != "" {
			clientAddr = client.Name + " (" + client.Address + ")"
		}
		v.clientTable.SetCell(row, 0, tview.NewTableCell(clientAddr))
		
		// Total duration (connection age in minutes)
		v.clientTable.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%.1f mins", client.TotalDuration)))
		
		// Idle time in seconds
		idleStr := fmt.Sprintf("%d s", client.Idle)
		if client.Idle > 60 {
			idleStr = fmt.Sprintf("%.1f mins", float64(client.Idle)/60.0)
		}
		v.clientTable.SetCell(row, 2, tview.NewTableCell(idleStr))
		
		// Last command
		lastCmd := client.LastCommand
		if lastCmd == "" {
			lastCmd = "none"
		}
		v.clientTable.SetCell(row, 3, tview.NewTableCell(lastCmd))
	}
}

// loadServerStats loads server statistics into the stats table
func (v *MonitorView) loadServerStats() {
	metrics, err := v.redis.GetMetrics()
	if err != nil {
		v.statsTable.Clear()
		v.statsTable.SetCell(0, 0, tview.NewTableCell("Error").SetTextColor(tcell.ColorRed))
		v.statsTable.SetCell(0, 1, tview.NewTableCell(fmt.Sprintf("Failed to load: %v", err)).SetTextColor(tcell.ColorRed))
		return
	}

	v.statsTable.Clear()
	
	// Calculate hit rate
	hitRate := float64(0)
	if metrics.KeyspaceHits+metrics.KeyspaceMisses > 0 {
		hitRate = float64(metrics.KeyspaceHits) / float64(metrics.KeyspaceHits+metrics.KeyspaceMisses) * 100
	}

	// Server statistics table
	stats := [][]string{
		{"Connected Clients", fmt.Sprintf("%d", metrics.ConnectedClients)},
		{"Used Memory", humanize.Bytes(uint64(metrics.UsedMemory))},
		{"Used Memory RSS", humanize.Bytes(uint64(metrics.UsedMemoryRss))},
		{"Total Commands", fmt.Sprintf("%d", metrics.TotalCommandsProcessed)},
		{"Ops/sec", fmt.Sprintf("%d", metrics.InstantaneousOpsPerSec)},
		{"Keyspace Hits", fmt.Sprintf("%d", metrics.KeyspaceHits)},
		{"Keyspace Misses", fmt.Sprintf("%d", metrics.KeyspaceMisses)},
		{"Hit Rate", fmt.Sprintf("%.2f%%", hitRate)},
		{"Uptime", getFormattedUptime(metrics.UptimeInSeconds)},
	}

	for i, stat := range stats {
		v.statsTable.SetCell(i, 0, tview.NewTableCell(stat[0]).SetTextColor(tcell.ColorGreen))
		v.statsTable.SetCell(i, 1, tview.NewTableCell(stat[1]).SetTextColor(tcell.ColorWhite))
	}
}

// loadSystemInfo loads system information
func (v *MonitorView) loadSystemInfo() {
	info, err := v.redis.Info()
	if err != nil {
		v.infoText.SetText(fmt.Sprintf("[red]Error loading Redis info: %s", err))
		return
	}

	timestamp := time.Now().Format("15:04:05")
	
	// Get cluster information
	clusterEnabled := getInfoValue(info, "cluster_enabled", "0")
	slowlogLen := getInfoValue(info, "slowlog_len", "0")
	totalConnections := getInfoValue(info, "total_connections_received", "0")
	rejectedConnections := getInfoValue(info, "rejected_connections", "0")
	
	infoText := fmt.Sprintf(`[yellow]Last Updated: %s[white]

[cyan]━━━ Connection Details ━━━[white]
  [green]Total Connections:[white] %s
  [green]Rejected Connections:[white] %s

[cyan]━━━ Performance ━━━[white]
  [green]Slow Log Length:[white] %s

[cyan]━━━ Cluster Info ━━━[white]
  [green]Cluster Enabled:[white] %s
`,
		timestamp,
		totalConnections,
		rejectedConnections,
		slowlogLen,
		clusterEnabled,
	)

	if clusterEnabled == "1" {
		infoText += "\n[cyan]━━━ Cluster Nodes ━━━[white]\n"
		infoText += v.getClusterNodesInfo()
	}

	// Add keyboard shortcuts help
	infoText += fmt.Sprintf(`

[cyan]━━━ Keyboard Shortcuts ━━━[white]
  [green]s/S:[white] Start/Stop monitoring
  [green]d/D:[white] Change refresh rate (%.0fs)
  [green]c/C:[white] Clear all tables
  [green]r/R:[white] Manual refresh
  [green]?:[white] Help
`, v.refreshRate.Seconds())

	v.infoText.SetText(infoText)
}

// getClusterNodesInfo returns formatted cluster nodes information for text display
func (v *MonitorView) getClusterNodesInfo() string {
	result, err := v.redis.ClusterNodes()
	if err != nil {
		return "  [red]Error getting cluster nodes info[white]"
	}

	lines := splitLines(result)
	if len(lines) == 0 {
		return "  [yellow]No cluster nodes found[white]"
	}

	info := ""
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 8 {
			continue
		}

		nodeID := parts[0][:8] + "..."
		hostPort := strings.Split(parts[1], "@")[0]
		flags := parts[2]
		linkState := parts[7]

		role := "slave"
		if strings.Contains(flags, "master") {
			role = "master"
		}
		if strings.Contains(flags, "myself") {
			role += " (self)"
		}

		statusColor := "green"
		if linkState != "connected" {
			statusColor = "red"
		}

		info += fmt.Sprintf("  [green]%s[white] - [yellow]%s[white] - [cyan]%s[white] - [%s]%s[white]\n", 
			nodeID, role, hostPort, statusColor, linkState)
	}

	return info
}

// Refresh refreshes the monitor view
func (v *MonitorView) Refresh() {
	v.loadData()
}

// getInfoValue safely extracts a value from info map
func getInfoValue(info map[string]interface{}, key, defaultValue string) string {
	if val, ok := info[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// getFormattedUptime formats uptime in seconds to human readable format with minutes
func getFormattedUptime(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm %ds", seconds/60, seconds%60)
	}
	if seconds < 86400 {
		return fmt.Sprintf("%dh %dm", seconds/3600, (seconds%3600)/60)
	}

	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := ((seconds % 86400) % 3600) / 60
	return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
}

// splitLines splits text into lines
func splitLines(text string) []string {
	if text == "" {
		return []string{}
	}

	lines := []string{}
	current := ""

	for _, char := range text {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

// joinLines joins lines with newlines
func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		result += line
		if i < len(lines)-1 {
			result += "\n"
		}
	}
	return result
}
