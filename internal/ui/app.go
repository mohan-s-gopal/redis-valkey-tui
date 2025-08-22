package ui

import (
	"fmt"
	"strings"

	"github.com/mohan-s-gopal/redis-valkey-tui/internal/config"
	"github.com/mohan-s-gopal/redis-valkey-tui/internal/logger"
	"github.com/mohan-s-gopal/redis-valkey-tui/internal/redis"

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
	HelpViewType
)

// App represents the main application
type App struct {
	app          *tview.Application
	pages        *tview.Pages
	contentPages *tview.Pages // For managing view switching
	redis        *redis.Client
	config       *config.Config

	// Views
	keysView    *KeysView
	infoView    *InfoView
	monitorView *MonitorView
	cliView     *CLIView
	configView  *ConfigView
	helpView    *HelpView

	// Current state
	currentView ViewType
	headerBar   *tview.Flex
	contextBar  *tview.TextView
	statusBar   *tview.TextView
	footerBar   *tview.TextView
	metrics     *Metrics

	// Help
	helpVisible bool
	helpModal   *tview.Modal

	// Testing flag
	testMode bool

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

	// Set initial view state directly without complex goroutines
	logger.Logger.Println("Setting up initial view state...")
	a.currentView = KeysViewType
	logger.Logger.Println("Initial view state configured")

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

	// Set up focus callback for KeysView
	a.keysView.SetFocusCallback(func(component tview.Primitive) {
		if !a.testMode && a.app != nil {
			a.app.SetFocus(component)
		}
	})

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

	logger.Logger.Println("Initializing HelpView...")
	if a.helpView = NewHelpView(); a.helpView == nil {
		return fmt.Errorf("failed to create HelpView")
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

	// Create footer with shortcuts
	a.footerBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetText("[yellow]Navigation:[white] 1=Keys 2=Info 3=Monitor 4=CLI 5=Config 6=Help | [yellow]Global:[white] ESC=home r=refresh ?=help Ctrl+C=quit")
	a.footerBar.SetBorder(true).
		SetTitle("Shortcuts").
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
	logger.Info("Setting up main layout...")

	// Create the main layout
	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow)
	logger.Debug("Created main flex layout")

	// Add header if it exists (fixed height of 3 lines for border + content + padding)
	if a.headerBar != nil {
		logger.Debug("Adding header bar to layout")
		mainLayout.AddItem(a.headerBar, 3, 0, false)
	} else {
		logger.Debug("No header bar to add")
	}

	// Create a content pages container to manage view switching
	logger.Debug("Creating content pages container")
	a.contentPages = tview.NewPages()

	// Add all views to the content pages
	logger.Debug("Adding all views to content pages")
	logger.Tracef("Adding Keys view: %p", a.keysView.GetComponent())
	a.contentPages.AddPage("keys", a.keysView.GetComponent(), true, true)

	logger.Tracef("Adding Info view: %p", a.infoView.GetComponent())
	a.contentPages.AddPage("info", a.infoView.GetComponent(), true, false)

	logger.Tracef("Adding Monitor view: %p", a.monitorView.GetComponent())
	a.contentPages.AddPage("monitor", a.monitorView.GetComponent(), true, false)

	logger.Tracef("Adding CLI view: %p", a.cliView.GetComponent())
	a.contentPages.AddPage("cli", a.cliView.GetComponent(), true, false)

	logger.Tracef("Adding Config view: %p", a.configView.GetComponent())
	a.contentPages.AddPage("config", a.configView.GetComponent(), true, false)

	logger.Tracef("Adding Help view: %p", a.helpView.GetComponent())
	a.contentPages.AddPage("help_view", a.helpView.GetComponent(), true, false)

	logger.Debug("All views added to content pages")

	// Add the content pages to the main layout
	logger.Debug("Adding content pages to main layout")
	mainLayout.AddItem(a.contentPages, 0, 1, true)

	// Add footer with shortcuts (fixed height of 3 lines)
	if a.footerBar != nil {
		logger.Debug("Adding footer bar to layout")
		mainLayout.AddItem(a.footerBar, 3, 0, false)
	} else {
		logger.Debug("No footer bar to add")
	}

	// Set up the pages
	logger.Debug("Setting up application pages")
	a.pages.RemovePage("main") // Remove the existing main page
	a.pages.AddPage("main", mainLayout, true, true)
	logger.Debug("Main page added to application pages")

	// Ensure help modal page exists
	if a.helpModal != nil && !a.pages.HasPage("help") {
		logger.Debug("Adding help modal page")
		a.pages.AddPage("help", a.helpModal, true, false)
	}

	// Set the root
	logger.Debug("Setting application root...")
	a.app.SetRoot(a.pages, true)

	// Set initial focus to current view
	if !a.testMode {
		currentView := a.getCurrentView()
		if currentView != nil {
			logger.Debugf("Setting initial focus to: %s", a.getViewName(a.currentView))
			a.app.SetFocus(currentView)
		} else {
			logger.Error("Current view is nil, cannot set initial focus")
		}
	} else {
		logger.Debug("Skipping focus setting (test mode)")
	}

	logger.Info("Application root set successfully")
}

// getCurrentView returns the current view component
func (a *App) getCurrentView() tview.Primitive {
	logger.Tracef("[getCurrentView] ENTRY: Getting current view for type: %s", a.getViewName(a.currentView))

	if a == nil {
		logger.Error("[getCurrentView] App instance is nil!")
		return nil
	}

	logger.Tracef("[getCurrentView] App instance is valid, calling getCurrentViewForType")
	result := a.getCurrentViewForType(a.currentView)

	if result == nil {
		logger.Error("[getCurrentView] Result is nil!")
	} else {
		logger.Tracef("[getCurrentView] SUCCESS: Returning view component %p for type: %s", result, a.getViewName(a.currentView))
	}

	logger.Tracef("[getCurrentView] EXIT")
	return result
}

// getCurrentViewForType returns the view component for the given view type
func (a *App) getCurrentViewForType(viewType ViewType) tview.Primitive {
	logger.Tracef("[getCurrentViewForType] ENTRY: Getting view for type: %s", a.getViewName(viewType))

	if a == nil {
		logger.Error("[getCurrentViewForType] App instance is nil!")
		return nil
	}

	var result tview.Primitive
	var viewName string

	switch viewType {
	case KeysViewType:
		viewName = "KeysView"
		logger.Tracef("[getCurrentViewForType] Case KeysViewType - checking a.keysView: %p", a.keysView)
		if a.keysView == nil {
			logger.Error("[getCurrentViewForType] keysView is nil!")
			return nil
		}
		logger.Tracef("[getCurrentViewForType] Calling keysView.GetComponent()")
		result = a.keysView.GetComponent()
		logger.Tracef("[getCurrentViewForType] keysView.GetComponent() returned: %p", result)

	case InfoViewType:
		viewName = "InfoView"
		logger.Tracef("[getCurrentViewForType] Case InfoViewType - checking a.infoView: %p", a.infoView)
		if a.infoView == nil {
			logger.Error("[getCurrentViewForType] infoView is nil!")
			return nil
		}
		logger.Tracef("[getCurrentViewForType] Calling infoView.GetComponent()")
		result = a.infoView.GetComponent()
		logger.Tracef("[getCurrentViewForType] infoView.GetComponent() returned: %p", result)

	case MonitorViewType:
		viewName = "MonitorView"
		logger.Tracef("[getCurrentViewForType] Case MonitorViewType - checking a.monitorView: %p", a.monitorView)
		if a.monitorView == nil {
			logger.Error("[getCurrentViewForType] monitorView is nil!")
			return nil
		}
		logger.Tracef("[getCurrentViewForType] Calling monitorView.GetComponent()")
		result = a.monitorView.GetComponent()
		logger.Tracef("[getCurrentViewForType] monitorView.GetComponent() returned: %p", result)

	case CLIViewType:
		viewName = "CLIView"
		logger.Tracef("[getCurrentViewForType] Case CLIViewType - checking a.cliView: %p", a.cliView)
		if a.cliView == nil {
			logger.Error("[getCurrentViewForType] cliView is nil!")
			return nil
		}
		logger.Tracef("[getCurrentViewForType] Calling cliView.GetComponent()")
		result = a.cliView.GetComponent()
		logger.Tracef("[getCurrentViewForType] cliView.GetComponent() returned: %p", result)

	case ConfigViewType:
		viewName = "ConfigView"
		logger.Tracef("[getCurrentViewForType] Case ConfigViewType - checking a.configView: %p", a.configView)
		if a.configView == nil {
			logger.Error("[getCurrentViewForType] configView is nil!")
			return nil
		}
		logger.Tracef("[getCurrentViewForType] Calling configView.GetComponent()")
		result = a.configView.GetComponent()
		logger.Tracef("[getCurrentViewForType] configView.GetComponent() returned: %p", result)

	case HelpViewType:
		viewName = "HelpView"
		logger.Tracef("[getCurrentViewForType] Case HelpViewType - checking a.helpView: %p", a.helpView)
		if a.helpView == nil {
			logger.Error("[getCurrentViewForType] helpView is nil!")
			return nil
		}
		logger.Tracef("[getCurrentViewForType] Calling helpView.GetComponent()")
		result = a.helpView.GetComponent()
		logger.Tracef("[getCurrentViewForType] helpView.GetComponent() returned: %p", result)

	default:
		viewName = "Default (KeysView)"
		logger.Warnf("[getCurrentViewForType] Unknown view type: %d, defaulting to KeysView", viewType)
		if a.keysView == nil {
			logger.Error("[getCurrentViewForType] Default keysView is nil!")
			return nil
		}
		logger.Tracef("[getCurrentViewForType] Calling default keysView.GetComponent()")
		result = a.keysView.GetComponent()
		logger.Tracef("[getCurrentViewForType] Default keysView.GetComponent() returned: %p", result)
	}

	if result == nil {
		logger.Errorf("[getCurrentViewForType] ERROR: %s GetComponent() returned nil!", viewName)
	} else {
		logger.Tracef("[getCurrentViewForType] SUCCESS: %s GetComponent() returned valid component: %p", viewName, result)
	}

	logger.Tracef("[getCurrentViewForType] EXIT: Returning %p for view type %s", result, a.getViewName(viewType))
	return result
}

// switchView switches to a different view
func (a *App) switchView(view ViewType) {
	logger.Infof("[switchView] ENTRY: Switching to view: %s", a.getViewName(view))
	logger.Debugf("[switchView] Previous view: %s, New view: %s", a.getViewName(a.currentView), a.getViewName(view))

	if a == nil {
		logger.Error("[switchView] CRITICAL: App instance is nil!")
		return
	}

	logger.Tracef("[switchView] Setting currentView from %s to %s", a.getViewName(a.currentView), a.getViewName(view))
	a.currentView = view
	logger.Tracef("[switchView] currentView set successfully to: %s", a.getViewName(a.currentView))

	// Command bar removed - no longer updating command bar text

	// Check prerequisites for UI operations
	logger.Tracef("[switchView] Checking UI operation prerequisites:")
	logger.Tracef("[switchView]   - testMode: %t", a.testMode)
	logger.Tracef("[switchView]   - app: %p", a.app)
	logger.Tracef("[switchView]   - contentPages: %p", a.contentPages)

	// Only do UI operations if not in test mode
	if !a.testMode && a.app != nil && a.contentPages != nil {
		logger.Debug("[switchView] All prerequisites met, performing UI operations")

		// Get the page name for the view
		logger.Tracef("[switchView] Getting page name for view: %s", a.getViewName(view))
		pageName := a.getPageNameForView(view)
		logger.Debugf("[switchView] Page name resolved: %s", pageName)

		// Get current view component
		logger.Tracef("[switchView] Getting current view component...")
		currentView := a.getCurrentView()
		logger.Tracef("[switchView] getCurrentView() completed, result: %p", currentView)

		if currentView != nil {
			logger.Tracef("[switchView] Current view is valid, attempting UI operations")

			// Try direct UI operations first (without QueueUpdateDraw)
			logger.Tracef("[switchView] Attempting direct UI operations (bypass QueueUpdateDraw)")

			// Direct page switch
			logger.Tracef("[switchView] Direct call to contentPages.SwitchToPage(%s)", pageName)
			a.contentPages.SwitchToPage(pageName)
			logger.Tracef("[switchView] Direct SwitchToPage completed")

			// Verify the direct page switch worked
			directFrontPage, directFrontName := a.contentPages.GetFrontPage()
			logger.Tracef("[switchView] Direct switch result - front page: %s (%s)", directFrontName, directFrontPage)

			// Direct focus
			logger.Tracef("[switchView] Direct call to app.SetFocus on component: %p", currentView)
			a.app.SetFocus(currentView)
			logger.Tracef("[switchView] Direct SetFocus completed")

			// Verify direct focus was set
			directFocusedComponent := a.app.GetFocus()
			logger.Tracef("[switchView] Direct focus result - focused component: %p", directFocusedComponent)

			logger.Tracef("[switchView] Direct UI operations completed successfully")

			// NOTE: QueueUpdateDraw removed due to hanging issues - direct operations are sufficient
			logger.Tracef("[switchView] Skipping QueueUpdateDraw to prevent hanging - direct operations handle view switching")

			// Log some diagnostics about the application state
			logger.Tracef("[switchView] Post-QueueUpdateDraw diagnostics:")
			logger.Tracef("[switchView]   - Application running: %p", a.app)
			logger.Tracef("[switchView]   - Current view type: %s", a.getViewName(a.currentView))
			logger.Tracef("[switchView]   - Content pages: %p", a.contentPages)
			if a.contentPages != nil {
				currentFrontPage, currentFrontName := a.contentPages.GetFrontPage()
				logger.Tracef("[switchView]   - Current front page: %s (%s)", currentFrontName, currentFrontPage)
			}
		} else {
			logger.Error("[switchView] CRITICAL: Current view is nil, cannot switch focus")
			logger.Tracef("[switchView] Debugging getCurrentView() call failure...")

			// Debug why getCurrentView returned nil
			logger.Tracef("[switchView] Debug: a.currentView = %d (%s)", a.currentView, a.getViewName(a.currentView))
			logger.Tracef("[switchView] Debug: Checking individual view pointers:")
			logger.Tracef("[switchView] Debug:   keysView: %p", a.keysView)
			logger.Tracef("[switchView] Debug:   infoView: %p", a.infoView)
			logger.Tracef("[switchView] Debug:   monitorView: %p", a.monitorView)
			logger.Tracef("[switchView] Debug:   cliView: %p", a.cliView)
			logger.Tracef("[switchView] Debug:   configView: %p", a.configView)
			logger.Tracef("[switchView] Debug:   helpView: %p", a.helpView)
		}
	} else {
		logger.Debug("[switchView] Skipping UI operations due to failed prerequisites:")
		if a.testMode {
			logger.Debug("[switchView]   - Reason: Test mode enabled")
		}
		if a.app == nil {
			logger.Error("[switchView]   - Reason: App is nil")
		}
		if a.contentPages == nil {
			logger.Error("[switchView]   - Reason: ContentPages is nil")
		}
	}

	logger.Infof("[switchView] EXIT: View switch completed for: %s", a.getViewName(view))
}

// switchViewForTest switches view without UI operations (for testing)
func (a *App) switchViewForTest(view ViewType) {
	logger.Logger.Printf("Switching to view (test): %s", a.getViewName(view))
	a.currentView = view
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
	case HelpViewType:
		return "Help"
	default:
		return "Unknown"
	}
}

// getPageNameForView returns the page name for a view type
func (a *App) getPageNameForView(view ViewType) string {
	logger.Tracef("[getPageNameForView] ENTRY: Getting page name for view type: %d (%s)", view, a.getViewName(view))

	var pageName string
	switch view {
	case KeysViewType:
		pageName = "keys"
	case InfoViewType:
		pageName = "info"
	case MonitorViewType:
		pageName = "monitor"
	case CLIViewType:
		pageName = "cli"
	case ConfigViewType:
		pageName = "config"
	case HelpViewType:
		pageName = "help_view"
	default:
		logger.Warnf("[getPageNameForView] Unknown view type: %d, defaulting to 'keys'", view)
		pageName = "keys"
	}

	logger.Tracef("[getPageNameForView] EXIT: Returning page name '%s' for view type %s", pageName, a.getViewName(view))
	return pageName
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

// handleGlobalKeys handles global key bindings
func (a *App) handleGlobalKeys(event *tcell.EventKey) *tcell.EventKey {
	logger.Tracef("Global key handler received event: Key=%v, Rune=%c, Mod=%v",
		event.Key(), event.Rune(), event.Modifiers())

	// Only handle specific global keys, let everything else pass through to views
	switch event.Key() {
	case tcell.KeyCtrlC:
		logger.Info("Ctrl+C pressed, shutting down application")
		a.cleanup()
		a.app.Stop()
		return nil
	case tcell.KeyEscape:
		logger.Info("ESC pressed, returning to main screen (Keys view)")
		a.switchView(KeysViewType)
		return nil
	case tcell.KeyCtrlR:
		logger.Info("Ctrl+R pressed, refreshing current view")
		a.refresh()
		return nil
	}

	// Handle number keys for quick view switching only if no input field has focus
	currentFocus := a.app.GetFocus()

	// Check if current focus is an InputField - if so, let numbers pass through
	if _, isInputField := currentFocus.(*tview.InputField); isInputField {
		return event // Pass through to input field
	}

	switch event.Rune() {
	case '1':
		logger.Debug("Number key '1' pressed, switching to Keys view")
		a.switchView(KeysViewType)
		return nil
	case '2':
		logger.Debug("Number key '2' pressed, switching to Info view")
		a.switchView(InfoViewType)
		return nil
	case '3':
		logger.Debug("Number key '3' pressed, switching to Monitor view")
		a.switchView(MonitorViewType)
		return nil
	case '4':
		logger.Debug("Number key '4' pressed, switching to CLI view")
		a.switchView(CLIViewType)
		return nil
	case '5':
		logger.Debug("Number key '5' pressed, switching to Config view")
		a.switchView(ConfigViewType)
		return nil
	case '6':
		logger.Debug("Number key '6' pressed, switching to Help view")
		if a.helpView != nil {
			a.switchView(HelpViewType)
		} else {
			logger.Warn("Help view is nil, cannot switch")
		}
		return nil
	case '?':
		logger.Debug("'?' key pressed, showing help modal")
		a.showHelp()
		return nil
	}

	// Let all other keys pass through to the views
	logger.Tracef("Key event passed through to view: Key=%v, Rune=%c", event.Key(), event.Rune())
	return event
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
	case "help":
		a.switchView(HelpViewType)
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
	return `redis-valkey-tui Help

Quick Navigation:
  1           Switch to Keys view
  2           Switch to Info view
  3           Switch to Monitor view
  4           Switch to CLI view
  5           Switch to Config view
  6           Switch to Help view

Navigation Commands:
  :keys       Switch to Keys view
  :info       Switch to Info view
  :monitor    Switch to Monitor view
  :cli        Switch to CLI view
  :config     Switch to Config view
  :help       Switch to Help view

Global Commands:
  :quit, :q   Quit application
  :refresh, :r Refresh current view
  Ctrl+C      Quit application
  Ctrl+R      Refresh current view
  ?           Show this help modal

Keys View:
  d           Delete selected key
  e           Edit selected key
  t           Set TTL for selected key
  /           Filter keys
  r           Refresh keys
  c           Execute command
  Enter       View key details

Info View:
  r           Refresh server info
  c           Execute command

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
  c           Execute command

Press ESC to close dialogs or return to previous view.`
}
