package ui

import (
	"fmt"
	"valkys/internal/redis"

	"github.com/rivo/tview"
)

// InfoView represents the server info view
type InfoView struct {
	redis     *redis.Client
	component *tview.TextView
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
	v.component = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)

	v.component.SetBorder(true).
		SetTitle("Server Information").
		SetBorderPadding(0, 0, 1, 1)
}

// GetComponent returns the main component
func (v *InfoView) GetComponent() tview.Primitive {
	return v.component
}

// loadInfo loads server information
func (v *InfoView) loadInfo() {
	info, err := v.redis.GetInfo()
	if err != nil {
		v.component.SetText(fmt.Sprintf("[red]Error loading server info: %s", err))
		return
	}

	// Format the info in a readable way
	text := "[yellow]Redis/Valkey Server Information[white]\n\n"

	// Key sections to highlight
	sections := []string{
		"redis_version",
		"redis_mode",
		"os",
		"arch_bits",
		"process_id",
		"uptime_in_seconds",
		"uptime_in_days",
		"tcp_port",
		"connected_clients",
		"used_memory_human",
		"used_memory_peak_human",
		"total_commands_processed",
		"instantaneous_ops_per_sec",
		"keyspace_hits",
		"keyspace_misses",
		"expired_keys",
		"evicted_keys",
		"keyspace_hit_rate",
	}

	// Display key metrics first
	for _, key := range sections {
		if value, exists := info[key]; exists {
			text += fmt.Sprintf("[green]%s:[white] %s\n", key, value)
		}
	}

	// Add separator
	text += "\n[yellow]All Server Information:[white]\n"

	// Display all other info
	for key, value := range info {
		// Skip if already shown above
		found := false
		for _, section := range sections {
			if key == section {
				found = true
				break
			}
		}
		if !found {
			text += fmt.Sprintf("[blue]%s:[white] %s\n", key, value)
		}
	}

	v.component.SetText(text)
}

// Refresh refreshes the info view
func (v *InfoView) Refresh() {
	v.loadInfo()
}
