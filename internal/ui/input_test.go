package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/mohan-s-gopal/redis-valkey-tui/internal/config"
	"github.com/mohan-s-gopal/redis-valkey-tui/internal/logger"
	"testing"
)

// TestInputHandlingAfterViewSwitch tests that input handling works after switching views
func TestInputHandlingAfterViewSwitch(t *testing.T) {
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
	app.testMode = true // Enable test mode to avoid UI operations
	app.setupUI()
	app.redis = newTestRedisClient()
	if app.redis == nil {
		t.Skip("Redis server not available for testing")
		return
	}

	err := app.initializeViews()
	assert.NoError(t, err, "View initialization should not error")

	// Test cases for each view transition
	testCases := []struct {
		name         string
		fromView     ViewType
		toView       ViewType
		testKey      rune
		shouldHandle bool
		description  string
	}{
		// Test global navigation keys work from each view
		{"Keys to Info", KeysViewType, InfoViewType, '2', true, "Should switch from Keys to Info"},
		{"Info to Monitor", InfoViewType, MonitorViewType, '3', true, "Should switch from Info to Monitor"},
		{"Monitor to CLI", MonitorViewType, CLIViewType, '4', true, "Should switch from Monitor to CLI"},
		{"CLI to Config", CLIViewType, ConfigViewType, '5', true, "Should switch from CLI to Config"},
		{"Config to Help", ConfigViewType, HelpViewType, '6', true, "Should switch from Config to Help"},
		{"Help to Keys", HelpViewType, KeysViewType, '1', true, "Should switch from Help to Keys"},

		// Test view-specific keys work after switching
		{"Keys view filter after switch", KeysViewType, KeysViewType, '/', true, "Should handle filter key in Keys view"},
		{"Keys view command after switch", KeysViewType, KeysViewType, 'c', true, "Should handle command key in Keys view"},
		{"Monitor view commands after switch", MonitorViewType, MonitorViewType, 's', true, "Should handle start/stop in Monitor view"},
		{"Info view refresh after switch", InfoViewType, InfoViewType, 'r', true, "Should handle refresh in Info view"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the starting view
			app.currentView = tc.fromView

			// Create a mock event for the test key
			event := tcell.NewEventKey(tcell.KeyRune, tc.testKey, tcell.ModNone)

			// Simulate the global key handler
			result := app.handleGlobalKeys(event)

			if tc.shouldHandle {
				if tc.testKey >= '1' && tc.testKey <= '6' {
					// Navigation keys should be consumed (return nil) and change view
					assert.Nil(t, result, "Navigation key should be consumed")
					assert.Equal(t, tc.toView, app.currentView, "View should change to expected view")
				} else {
					// View-specific keys should pass through (return event)
					assert.Equal(t, event, result, "View-specific key should pass through to view")
				}
			}
		})
	}
}

// TestViewInputCaptureChain tests that input capture chain works correctly
func TestViewInputCaptureChain(t *testing.T) {
	logger.Init()

	cfg := &config.Config{
		Redis: config.RedisConfig{Host: "localhost", Port: 6379, DB: 0},
		UI:    config.UIConfig{MaxKeys: 100},
	}

	app := NewApp(cfg)
	app.testMode = true
	app.setupUI()
	app.redis = newTestRedisClient()
	if app.redis == nil {
		t.Skip("Redis server not available for testing")
		return
	}

	err := app.initializeViews()
	assert.NoError(t, err)

	// Test that each view has proper input capture set up
	testViews := []struct {
		viewType ViewType
		viewName string
	}{
		{KeysViewType, "Keys"},
		{InfoViewType, "Info"},
		{MonitorViewType, "Monitor"},
		{CLIViewType, "CLI"},
		{ConfigViewType, "Config"},
		{HelpViewType, "Help"},
	}

	for _, tv := range testViews {
		t.Run(tv.viewName+"_view_input_capture", func(t *testing.T) {
			app.currentView = tv.viewType
			currentView := app.getCurrentView()
			assert.NotNil(t, currentView, "Current view should not be nil")

			// Each view should be able to handle input
			// This test ensures the view components exist and are properly configured
		})
	}
}

// TestGlobalKeyHandling tests that global keys are handled correctly
func TestGlobalKeyHandling(t *testing.T) {
	logger.Init()

	cfg := &config.Config{
		Redis: config.RedisConfig{Host: "localhost", Port: 6379, DB: 0},
		UI:    config.UIConfig{MaxKeys: 100},
	}

	app := NewApp(cfg)
	app.testMode = true
	app.setupUI()
	app.redis = newTestRedisClient()
	if app.redis == nil {
		t.Skip("Redis server not available for testing")
		return
	}

	err := app.initializeViews()
	assert.NoError(t, err)

	globalKeys := []struct {
		key      tcell.Key
		rune     rune
		expected string
	}{
		{tcell.KeyRune, '1', "should switch to Keys view"},
		{tcell.KeyRune, '2', "should switch to Info view"},
		{tcell.KeyRune, '3', "should switch to Monitor view"},
		{tcell.KeyRune, '4', "should switch to CLI view"},
		{tcell.KeyRune, '5', "should switch to Config view"},
		{tcell.KeyRune, '6', "should switch to Help view"},
		{tcell.KeyRune, '?', "should show help"},
		{tcell.KeyCtrlR, 0, "should refresh"},
	}

	for _, gk := range globalKeys {
		t.Run(gk.expected, func(t *testing.T) {
			event := tcell.NewEventKey(gk.key, gk.rune, tcell.ModNone)
			result := app.handleGlobalKeys(event)

			if gk.rune >= '1' && gk.rune <= '6' {
				// Navigation keys should be consumed
				assert.Nil(t, result, "Navigation key should be consumed")
			} else if gk.rune == '?' || gk.key == tcell.KeyCtrlR {
				// Help and refresh should be consumed
				assert.Nil(t, result, "Global command should be consumed")
			}
		})
	}
}

// TestFocusManagement tests that focus is properly managed during view switches
func TestFocusManagement(t *testing.T) {
	logger.Init()

	cfg := &config.Config{
		Redis: config.RedisConfig{Host: "localhost", Port: 6379, DB: 0},
		UI:    config.UIConfig{MaxKeys: 100},
	}

	app := NewApp(cfg)
	app.testMode = true
	app.setupUI()
	app.redis = newTestRedisClient()
	if app.redis == nil {
		t.Skip("Redis server not available for testing")
		return
	}

	err := app.initializeViews()
	assert.NoError(t, err)

	// Test focus management during view switches
	views := []ViewType{KeysViewType, InfoViewType, MonitorViewType, CLIViewType, ConfigViewType, HelpViewType}

	for _, view := range views {
		t.Run(app.getViewName(view)+"_focus", func(t *testing.T) {
			app.currentView = view
			currentView := app.getCurrentView()
			assert.NotNil(t, currentView, "Current view should not be nil for "+app.getViewName(view))

			// Test that we can get the component for focus setting
			assert.Implements(t, (*tview.Primitive)(nil), currentView, "View should implement tview.Primitive")
		})
	}
}
