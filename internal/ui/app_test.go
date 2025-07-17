package ui

import (
	"context"
	"testing"
	"redis-cli-dashboard/internal/config"
	"redis-cli-dashboard/internal/redis"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// mockRedisClient is a mock implementation for testing
type mockRedisClient struct {
	infoResult map[string]interface{}
	infoErr    error
}

// Create a custom Redis client for testing
func newTestRedisClient() *redis.Client {
	// Create a mock Redis client with minimal functionality
	mockRDB := goredis.NewClient(&goredis.Options{})
	return &redis.Client{
		rdb: mockRDB,
		ctx: context.Background(),
	}
}

// TestNewApp tests the creation of a new App instance
func TestNewApp(t *testing.T) {
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
	assert.NotNil(t, app.commandBar, "Command bar should be initialized")
	assert.NotNil(t, app.helpModal, "Help modal should be initialized")

	// Test view initialization with mock Redis client
	app.redis = newTestRedisClient()

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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app.switchView(tc.viewType)
			assert.Equal(t, tc.viewType, app.currentView, "View should be switched correctly")
		})
	}
}

// TestCleanup tests the cleanup process
func TestCleanup(t *testing.T) {
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
	
	// Create a metrics stop channel
	app.metricsStopChan = make(chan struct{})

	// Test cleanup
	app.cleanup()

	// Verify that the metrics stop channel is closed
	select {
	case _, ok := <-app.metricsStopChan:
		assert.False(t, ok, "Metrics stop channel should be closed")
	default:
		t.Error("Metrics stop channel should be closed")
	}
}

