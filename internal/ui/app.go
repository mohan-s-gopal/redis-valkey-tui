package ui

import (
	"fmt"
	"strings"

	"redis-cli-dashboard/internal/config"
	"redis-cli-dashboard/internal/redis"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ViewType represents different view types
type ViewType int

const (
	KeysViewType ViewType = iota
	InfoViewType
	MonitorViewType
	CLIViewType
	ConfigViewType
)

// App represents the main application
type App struct {
	app    *tview.Application
	pages  *tview.Pages
	redis  *redis.Client
	config *config.Config

	// Views
	keysView    *KeysView
	infoView    *InfoView
	monitorView *MonitorView
	cliView     *CLIView
	configView  *ConfigView

	// Current state
	currentView ViewType
	statusBar   *tview.TextView
	commandBar  *tview.TextView

	// Help
	helpVisible bool
	helpModal   *tview.Modal
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config) *App {
	app := &App{
		app:         tview.NewApplication(),
		pages:       tview.NewPages(),
		config:      cfg,
		currentView: KeysViewType,
	}

	app.setupUI()
	return app
}

// Run starts the application
func (a *App) Run() error {
	// Connect to Redis
	redisClient, err := redis.New(&a.config.Redis)
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	a.redis = redisClient

	// Initialize views
	a.keysView = NewKeysView(a.redis, a.config)
	a.infoView = NewInfoView(a.redis)
	a.monitorView = NewMonitorView(a.redis)
	a.cliView = NewCLIView(a.redis)
	a.configView = NewConfigView(a.config)

	// Setup main layout
	a.setupLayout()

	// Set initial view
	a.switchView(KeysViewType)

	// Start the application
	return a.app.Run()
}

// setupUI initializes the UI components
func (a *App) setupUI() {
	// Create status bar
	a.statusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetText("Ready")
	a.statusBar.SetBorder(true).
		SetTitle("Status").
		SetBorderPadding(0, 0, 1, 1)

	// Create command bar
	a.commandBar = tview.NewTextView().
		SetDynamicColors(true).
		SetText("Press ':' to enter command mode, '?' for help, 'Ctrl+C' to quit")
	a.commandBar.SetBorder(true).
		SetTitle("Commands").
		SetBorderPadding(0, 0, 1, 1)

	// Create help modal
	a.helpModal = tview.NewModal().
		SetText(a.getHelpText()).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.hideHelp()
		})

	// Set up global key bindings
	a.app.SetInputCapture(a.handleGlobalKeys)
}

// setupLayout creates the main layout
func (a *App) setupLayout() {
	// Create main flex container
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.getCurrentView(), 0, 1, true).
		AddItem(a.statusBar, 3, 0, false).
		AddItem(a.commandBar, 3, 0, false)

	a.pages.AddPage("main", flex, true, true)
	a.pages.AddPage("help", a.helpModal, true, false)

	a.app.SetRoot(a.pages, true)
}

// getCurrentView returns the current view component
func (a *App) getCurrentView() tview.Primitive {
	switch a.currentView {
	case KeysViewType:
		return a.keysView.GetComponent()
	case InfoViewType:
		return a.infoView.GetComponent()
	case MonitorViewType:
		return a.monitorView.GetComponent()
	case CLIViewType:
		return a.cliView.GetComponent()
	case ConfigViewType:
		return a.configView.GetComponent()
	default:
		return a.keysView.GetComponent()
	}
}

// switchView switches to a different view
func (a *App) switchView(view ViewType) {
	a.currentView = view

	// Update status bar
	viewName := a.getViewName(view)
	a.statusBar.SetText(fmt.Sprintf("[green]%s view[white] - %s", viewName, a.getViewStatus()))

	// Update command bar
	a.commandBar.SetText(a.getViewCommands())

	// Refresh layout
	a.setupLayout()

	// Focus the new view
	a.app.SetFocus(a.getCurrentView())
}

// getViewName returns the name of the view
func (a *App) getViewName(view ViewType) string {
	switch view {
	case KeysViewType:
		return "Keys"
	case InfoViewType:
		return "Info"
	case MonitorViewType:
		return "Monitor"
	case CLIViewType:
		return "CLI"
	case ConfigViewType:
		return "Config"
	default:
		return "Unknown"
	}
}

// getViewStatus returns the status for the current view
func (a *App) getViewStatus() string {
	switch a.currentView {
	case KeysViewType:
		return fmt.Sprintf("DB:%d", a.config.Redis.DB)
	case InfoViewType:
		return "Server information"
	case MonitorViewType:
		return "Real-time monitoring"
	case CLIViewType:
		return "Redis CLI"
	case ConfigViewType:
		return "Configuration"
	default:
		return "Ready"
	}
}

// getViewCommands returns the commands for the current view
func (a *App) getViewCommands() string {
	switch a.currentView {
	case KeysViewType:
		return "d=delete e=edit t=ttl /=filter r=refresh :=command ?=help"
	case InfoViewType:
		return "r=refresh :=command ?=help"
	case MonitorViewType:
		return "s=start/stop c=clear r=refresh :=command ?=help"
	case CLIViewType:
		return "Enter=execute ↑↓=history Ctrl+L=clear :=command ?=help"
	case ConfigViewType:
		return "s=save r=reset :=command ?=help"
	default:
		return "?=help"
	}
}

// handleGlobalKeys handles global key bindings
func (a *App) handleGlobalKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlC:
		a.app.Stop()
		return nil
	case tcell.KeyCtrlR:
		a.refresh()
		return nil
	}

	switch event.Rune() {
	case ':':
		a.enterCommandMode()
		return nil
	case '?':
		a.showHelp()
		return nil
	}

	return event
}

// enterCommandMode enters command mode
func (a *App) enterCommandMode() {
	var input *tview.InputField

	input = tview.NewInputField().
		SetLabel("Command: ").
		SetFieldWidth(50).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				command := input.GetText()
				a.executeCommand(command)
				a.pages.SwitchToPage("main")
			} else if key == tcell.KeyEscape {
				a.pages.SwitchToPage("main")
			}
		})

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(input, 54, 0, true).
			AddItem(nil, 0, 1, false), 3, 0, true).
		AddItem(nil, 0, 1, false)

	a.pages.AddPage("command", flex, true, true)
	a.pages.SwitchToPage("command")
	a.app.SetFocus(input)
}

// executeCommand executes a command
func (a *App) executeCommand(command string) {
	command = strings.TrimSpace(command)
	if command == "" {
		return
	}

	switch command {
	case "keys":
		a.switchView(KeysViewType)
	case "info":
		a.switchView(InfoViewType)
	case "monitor":
		a.switchView(MonitorViewType)
	case "cli":
		a.switchView(CLIViewType)
	case "config":
		a.switchView(ConfigViewType)
	case "quit", "q":
		a.app.Stop()
	case "refresh", "r":
		a.refresh()
	default:
		a.statusBar.SetText(fmt.Sprintf("[red]Unknown command: %s", command))
	}
}

// showHelp shows the help modal
func (a *App) showHelp() {
	a.helpVisible = true
	a.pages.SwitchToPage("help")
}

// hideHelp hides the help modal
func (a *App) hideHelp() {
	a.helpVisible = false
	a.pages.SwitchToPage("main")
}

// refresh refreshes the current view
func (a *App) refresh() {
	switch a.currentView {
	case KeysViewType:
		a.keysView.Refresh()
	case InfoViewType:
		a.infoView.Refresh()
	case MonitorViewType:
		a.monitorView.Refresh()
	case CLIViewType:
		a.cliView.Refresh()
	case ConfigViewType:
		a.configView.Refresh()
	}

	a.statusBar.SetText(fmt.Sprintf("[green]%s view[white] - Refreshed", a.getViewName(a.currentView)))
}

// getHelpText returns the help text
func (a *App) getHelpText() string {
	return `redis-cli-dashboard Help

Navigation:
  :keys       Switch to Keys view
  :info       Switch to Info view
  :monitor    Switch to Monitor view
  :cli        Switch to CLI view
  :config     Switch to Config view

Global Commands:
  :quit, :q   Quit application
  :refresh, :r Refresh current view
  Ctrl+C      Quit application
  Ctrl+R      Refresh current view
  ?           Show this help

Keys View:
  d           Delete selected key
  e           Edit selected key
  t           Set TTL for selected key
  /           Filter keys
  r           Refresh keys
  Enter       View key details

Info View:
  r           Refresh server info

Monitor View:
  s           Start/stop monitoring
  c           Clear screen
  r           Refresh metrics

CLI View:
  Enter       Execute command
  ↑/↓         Navigate command history
  Ctrl+L      Clear screen

Config View:
  s           Save configuration
  r           Reset to defaults

Press ESC to close dialogs or return to previous view.`
}
