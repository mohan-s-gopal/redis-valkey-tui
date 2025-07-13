package ui

import (
	"fmt"
	"valkys/internal/config"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ConfigView represents the configuration view
type ConfigView struct {
	config    *config.Config
	component *tview.TextView
}

// NewConfigView creates a new config view
func NewConfigView(cfg *config.Config) *ConfigView {
	view := &ConfigView{
		config: cfg,
	}

	view.setupUI()
	view.loadConfig()

	return view
}

// setupUI initializes the UI components
func (v *ConfigView) setupUI() {
	v.component = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)

	v.component.SetInputCapture(v.handleInput)

	v.component.SetBorder(true).
		SetTitle("Configuration").
		SetBorderPadding(0, 0, 1, 1)
}

// GetComponent returns the main component
func (v *ConfigView) GetComponent() tview.Primitive {
	return v.component
}

// handleInput handles input for the config view
func (v *ConfigView) handleInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 's':
		v.saveConfig()
		return nil
	case 'r':
		v.resetConfig()
		return nil
	}

	return event
}

// loadConfig loads and displays the configuration
func (v *ConfigView) loadConfig() {
	text := `[yellow]Valkys Configuration[white]

[green]Redis Connection:[white]
  Host: [cyan]%s[white]
  Port: [cyan]%d[white]
  Password: [cyan]%s[white]
  Database: [cyan]%d[white]
  Timeout: [cyan]%d ms[white]
  Pool Size: [cyan]%d[white]

[green]UI Settings:[white]
  Theme: [cyan]%s[white]
  Refresh Interval: [cyan]%d ms[white]
  Max Keys: [cyan]%d[white]
  Show Memory: [cyan]%t[white]
  Show TTL: [cyan]%t[white]

[yellow]Commands:[white]
  s - Save configuration
  r - Reset to defaults
  ? - Show help

[gray]Note: Configuration editing is not yet implemented.
Use the config file directly for now.[white]`

	password := v.config.Redis.Password
	if password != "" {
		password = "***"
	}

	formattedText := fmt.Sprintf(text,
		v.config.Redis.Host,
		v.config.Redis.Port,
		password,
		v.config.Redis.DB,
		v.config.Redis.Timeout,
		v.config.Redis.PoolSize,
		v.config.UI.Theme,
		v.config.UI.RefreshInterval,
		v.config.UI.MaxKeys,
		v.config.UI.ShowMemory,
		v.config.UI.ShowTTL,
	)

	v.component.SetText(formattedText)
}

// saveConfig saves the configuration
func (v *ConfigView) saveConfig() {
	err := v.config.Save()
	if err != nil {
		v.component.SetText(fmt.Sprintf("[red]Error saving config: %s", err))
	} else {
		v.component.SetText("[green]Configuration saved successfully!")
	}
}

// resetConfig resets the configuration to defaults
func (v *ConfigView) resetConfig() {
	defaultConfig := config.Default()
	*v.config = *defaultConfig
	v.loadConfig()
}

// Refresh refreshes the config view
func (v *ConfigView) Refresh() {
	v.loadConfig()
}
