package ui

import (
	"fmt"
	"strings"
	"time"

	"redis-cli-dashboard/internal/config"
	"redis-cli-dashboard/internal/logger"
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
	headerBar   *tview.Flex
	contextBar  *tview.TextView
	statusBar   *tview.TextView
	commandBar  *tview.TextView
	metrics     *Metrics

	// Help
	helpVisible bool
	helpModal   *tview.Modal

	// Metrics update control
	metricsStopChan chan struct{}
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config) *App {
	app := &App{
		app:         tview.NewApplication(),
		pages:       tview.NewPages(),
		config:      cfg,
		metrics:     NewMetrics(),
		currentView: KeysViewType,
	}

	return app
}

// Run starts the application
func (a *App) Run() error {
	logger.Logger.Println("Starting application with configuration:", 
		fmt.Sprintf("Redis: %s:%d/DB%d, MaxKeys: %d", 
			a.config.Redis.Host, 
			a.config.Redis.Port, 
			a.config.Redis.DB,
			a.config.UI.MaxKeys))

	// Validate configuration
	if a.config == nil {
		return fmt.Errorf("invalid configuration: config is nil")
	}

	// Connect to Redis with timeout
	logger.Logger.Printf("Establishing Redis connection to %s:%d...", a.config.Redis.Host, a.config.Redis.Port)
	redisClient, err := redis.New(&a.config.Redis)
	if err != nil {
		logger.Logger.Printf("CRITICAL: Redis connection failed: %v", err)
		return fmt.Errorf("redis connection failed: %w", err)
	}

	// Get Redis server info
	info, err := redisClient.Info()
	if err != nil {
		logger.Logger.Printf("CRITICAL: Failed to get Redis server info: %v", err)
		return fmt.Errorf("failed to get Redis server info: %w", err)
	}
	
	version, _ := info["redis_version"].(string)
	if version == "" {
		version = "unknown"
	}
	
	logger.Logger.Printf("SUCCESS: Connected to Redis server version: %s", version)
	a.redis = redisClient

	// Initialize UI components with error handling
	logger.Logger.Println("Initializing UI components...")
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Printf("PANIC in UI initialization: %v", r)
			panic(r) // Re-panic after logging
		}
	}()
	
	a.setupUI()
	logger.Logger.Println("UI components initialized successfully")

	// Initialize views with detailed logging
	logger.Logger.Println("Initializing application views...")
	if err := a.initializeViews(); err != nil {
		logger.Logger.Printf("ERROR: Failed to initialize views: %v", err)
		return fmt.Errorf("view initialization failed: %w", err)
	}

	// Create and initialize header
	logger.Logger.Println("Creating application header...")
	if a.headerBar = a.createHeader(); a.headerBar == nil {
		logger.Logger.Println("ERROR: Failed to create header bar")
		return fmt.Errorf("header creation failed")
	}
	logger.Logger.Println("Header created and initialized")

	// Setup initial layout
	logger.Logger.Println("Configuring application layout...")
	a.setupLayout()
	logger.Logger.Println("Application layout configured successfully")

	// Initialize view with error handling and timeout
	logger.Logger.Println("Setting up initial view state...")
	viewInitialized := make(chan bool, 1)
	
	// Setup initial view in a separate goroutine
	go func() {
		a.app.QueueUpdateDraw(func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Logger.Printf("PANIC in view initialization: %v", r)
					viewInitialized <- false
					return
				}
				viewInitialized <- true
			}()
			logger.Logger.Println("Switching to initial view (KeysView)...")
			a.switchView(KeysViewType)
			logger.Logger.Println("Initial view initialized successfully")
		})
	}()

	// Wait for view initialization with timeout
	select {
	case success := <-viewInitialized:
		if !success {
			return fmt.Errorf("failed to initialize application view")
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("view initialization timed out after 5 seconds")
	}

	// Start the application with error handling
	logger.Logger.Println("Starting application main loop...")
	if err := a.app.Run(); err != nil {
		logger.Logger.Printf("CRITICAL: Application terminated with error: %v", err)
		return fmt.Errorf("application runtime error: %w", err)
	}
	
	logger.Logger.Println("Application terminated normally")
	return nil
}

// initializeViews initializes all application views
func (a *App) initializeViews() error {
	if a.redis == nil {
		return fmt.Errorf("cannot initialize views: Redis client is nil")
	}

	// Initialize each view with error checking
	logger.Logger.Println("Initializing KeysView...")
	if a.keysView = NewKeysView(a.redis, a.config); a.keysView == nil {
		return fmt.Errorf("failed to create KeysView")
	}

	logger.Logger.Println("Initializing InfoView...")
	if a.infoView = NewInfoView(a.redis); a.infoView == nil {
		return fmt.Errorf("failed to create InfoView")
	}

	logger.Logger.Println("Initializing MonitorView...")
	if a.monitorView = NewMonitorView(a.redis); a.monitorView == nil {
		return fmt.Errorf("failed to create MonitorView")
	}

	logger.Logger.Println("Initializing CLIView...")
	if a.cliView = NewCLIView(a.redis); a.cliView == nil {
		return fmt.Errorf("failed to create CLIView")
	}

	logger.Logger.Println("Initializing ConfigView...")
	if a.configView = NewConfigView(a.config); a.configView == nil {
		return fmt.Errorf("failed to create ConfigView")
	}

	logger.Logger.Println("All views initialized successfully")
	return nil
}

// setupUI initializes the UI components
func (a *App) setupUI() {
	// Create status bar
	a.statusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetText("Ready")

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
	logger.Logger.Println("Setting up main layout...")
	
	// Create the main layout
	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow)

	// Add header if it exists
	if a.headerBar != nil {
		mainLayout.AddItem(a.headerBar, 1, 0, false)
	}

	// Create main content area
	currentView := a.getCurrentView()
	if currentView != nil {
		mainLayout.AddItem(currentView, 0, 1, true)
	}

	// Add command bar if it exists
	if a.commandBar != nil {
		mainLayout.AddItem(a.commandBar, 1, 0, false)
	}

	// Set up the pages
	a.pages.RemovePage("main")  // Remove the existing main page
	a.pages.AddPage("main", mainLayout, true, true)
	
	// Ensure help modal page exists
	if a.helpModal != nil && !a.pages.HasPage("help") {
		a.pages.AddPage("help", a.helpModal, true, false)
	}

	// Set the root
	logger.Logger.Println("Setting application root...")
	a.app.SetRoot(a.pages, true)
	logger.Logger.Println("Application root set")
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
	logger.Logger.Printf("Switching to view: %s", a.getViewName(view))
	a.currentView = view

	// Update command bar text
	if a.commandBar != nil {
		logger.Logger.Println("Updating command bar text")
		a.commandBar.SetText(a.getViewCommands())
	}

	// Queue the layout update
	a.app.QueueUpdateDraw(func() {
		// Refresh layout
		a.setupLayout()

		// Focus the new view
		currentView := a.getCurrentView()
		if currentView != nil {
			a.app.SetFocus(currentView)
		}
	})
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
		a.cleanup()
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
		a.cleanup()
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

// cleanup performs necessary cleanup before application exit
func (a *App) cleanup() {
	logger.Logger.Println("Performing application cleanup...")

	// Stop metrics collection
	if a.metricsStopChan != nil {
		close(a.metricsStopChan)
		a.metricsStopChan = nil
	}

	// Close Redis connection
	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			logger.Logger.Printf("Error closing Redis connection: %v", err)
		}
	}

	logger.Logger.Println("Cleanup completed")
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
