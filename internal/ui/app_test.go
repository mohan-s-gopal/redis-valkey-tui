package ui

import (
	"testing"
	"redis-cli-dashboard/internal/config"
	"redis-cli-dashboard/internal/redis"
	"redis-cli-dashboard/internal/logger"
	"github.com/stretchr/testify/assert"
)

// Create a custom Redis client for testing
func newTestRedisClient() *redis.Client {
	// For testing, we'll create a minimal config that actually connects to a test Redis
	// If no Redis is available, we'll skip the tests that require it
	cfg := &config.RedisConfig{
		Host: "localhost",
		Port: 6379, // Use standard Redis port
		DB:   0,
	}
	
	// Try to create a real connection, but don't fail if it's not available
	client, err := redis.New(cfg)
	if err != nil {
		// Return nil if Redis is not available - tests will handle this
		return nil
	}
	return client
}

// TestNewApp tests the creation of a new App instance
func TestNewApp(t *testing.T) {
	// Initialize logger for tests
	logger.Init()
	
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
		UI: config.UIConfig{
			MaxKeys: 100,
		},
	}

	app := NewApp(cfg)
	assert.NotNil(t, app, "App should not be nil")
	assert.NotNil(t, app.app, "tview.Application should not be nil")
	assert.NotNil(t, app.pages, "Pages should not be nil")
	assert.NotNil(t, app.config, "Config should not be nil")
	assert.Equal(t, KeysViewType, app.currentView, "Initial view should be KeysView")
}

// TestAppInitialization tests the initialization process
func TestAppInitialization(t *testing.T) {
	// Initialize logger for tests
	logger.Init()
	
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
		UI: config.UIConfig{
			MaxKeys: 100,
		},
	}

	app := NewApp(cfg)
	
	// Test setupUI
	app.setupUI()
	assert.NotNil(t, app.statusBar, "Status bar should be initialized")
	assert.NotNil(t, app.helpModal, "Help modal should be initialized")

	// Test view initialization with mock Redis client
	app.redis = newTestRedisClient()
	if app.redis == nil {
		t.Skip("Redis server not available for testing")
		return
	}

	err := app.initializeViews()
	assert.NoError(t, err, "View initialization should not error")
	assert.NotNil(t, app.keysView, "KeysView should be initialized")
	assert.NotNil(t, app.infoView, "InfoView should be initialized")
	assert.NotNil(t, app.monitorView, "MonitorView should be initialized")
	assert.NotNil(t, app.cliView, "CLIView should be initialized")
	assert.NotNil(t, app.configView, "ConfigView should be initialized")
}

// TestViewSwitching tests the view switching functionality
func TestViewSwitching(t *testing.T) {
	// Initialize logger for tests
	logger.Init()
	
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
		UI: config.UIConfig{
			MaxKeys: 100,
		},
	}

	app := NewApp(cfg)
	app.setupUI()
	app.redis = newTestRedisClient()
	if app.redis == nil {
		t.Skip("Redis server not available for testing")
		return
	}
	
	err := app.initializeViews()
	assert.NoError(t, err, "View initialization should not error")
	
	testCases := []struct {
		name     string
		viewType ViewType
	}{
		{"Switch to KeysView", KeysViewType},
		{"Switch to InfoView", InfoViewType},
		{"Switch to MonitorView", MonitorViewType},
		{"Switch to CLIView", CLIViewType},
		{"Switch to ConfigView", ConfigViewType},
		{"Switch to HelpView", HelpViewType},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Enable test mode before switching views
			app.testMode = true
			app.switchView(tc.viewType)
			assert.Equal(t, tc.viewType, app.currentView, "View should be switched correctly")
			
			// Test that getCurrentView returns the right component
			currentView := app.getCurrentView()
			assert.NotNil(t, currentView, "Current view should not be nil")
		})
	}
}

// TestViewSwitchingLogic tests the view switching logic without UI operations
func TestViewSwitchingLogic(t *testing.T) {
	// Initialize logger for tests
	logger.Init()
	
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
		UI: config.UIConfig{
			MaxKeys: 100,
		},
	}

	app := NewApp(cfg)
	app.testMode = true  // Enable test mode to avoid UI operations
	app.setupUI()
	app.redis = newTestRedisClient()
	if app.redis == nil {
		t.Skip("Redis server not available for testing")
		return
	}
	
	err := app.initializeViews()
	assert.NoError(t, err, "View initialization should not error")
	
	// Test getViewName method
	assert.Equal(t, "Keys", app.getViewName(KeysViewType))
	assert.Equal(t, "Info", app.getViewName(InfoViewType))
	assert.Equal(t, "Monitor", app.getViewName(MonitorViewType))
	assert.Equal(t, "CLI", app.getViewName(CLIViewType))
	assert.Equal(t, "Config", app.getViewName(ConfigViewType))
	assert.Equal(t, "Help", app.getViewName(HelpViewType))
	
	// Test executeCommand method
	app.executeCommand("keys")
	assert.Equal(t, KeysViewType, app.currentView)
	
	app.executeCommand("info")
	assert.Equal(t, InfoViewType, app.currentView)
	
	app.executeCommand("monitor")
	assert.Equal(t, MonitorViewType, app.currentView)
	
	app.executeCommand("cli")
	assert.Equal(t, CLIViewType, app.currentView)
	
	app.executeCommand("config")
	assert.Equal(t, ConfigViewType, app.currentView)
}

// TestCleanup tests the cleanup process
func TestCleanup(t *testing.T) {
	// Initialize logger for tests
	logger.Init()
	
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
		UI: config.UIConfig{
			MaxKeys: 100,
		},
	}

	app := NewApp(cfg)
	app.redis = newTestRedisClient()
	
	// Create a metrics stop channel (simulating what would happen in real app)
	app.metricsStopChan = make(chan struct{})

	// Test cleanup
	app.cleanup()

	// Verify that the metrics stop channel is closed and set to nil
	assert.Nil(t, app.metricsStopChan, "Metrics stop channel should be set to nil after cleanup")
}

