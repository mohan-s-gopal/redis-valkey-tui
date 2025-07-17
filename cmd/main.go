package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"

	"redis-cli-dashboard/internal/config"
	"redis-cli-dashboard/internal/logger"
	"redis-cli-dashboard/internal/ui"
)

// Main is the main entry point
func Main() {
	// Initialize logger
	if err := logger.Init(); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()
	
	logger.Logger.Println("Starting redis-cli-dashboard...")
	var (
		host     = flag.String("host", "", "Redis host")
		port     = flag.Int("port", 0, "Redis port")
		password = flag.String("password", "", "Redis password")
		db       = flag.Int("db", -1, "Redis database number")
		help     = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override with command line flags
	if *host != "" {
		cfg.Redis.Host = *host
	}
	if *port != 0 {
		cfg.Redis.Port = *port
	}
	if *password != "" {
		cfg.Redis.Password = *password
	}
	if *db != -1 {
		cfg.Redis.DB = *db
	}

	// Create and run the application
	app := ui.NewApp(cfg)
	if err := app.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func showHelp() {
	fmt.Print(`redis-cli-dashboard - A k9s-inspired TUI client for Redis/Valkey

Usage: redis-cli-dashboard [options]

Options:
  -host string
        Redis host (default: localhost)
  -port int
        Redis port (default: 6379)
  -password string
        Redis password
  -db int
        Redis database number (default: 0)
  -help
        Show this help

Navigation:
  :keys       Switch to Keys view
  :info       Switch to Info view
  :monitor    Switch to Monitor view
  :cli        Switch to CLI view
  :config     Switch to Config view
  
  /           Filter/search
  Ctrl+R      Refresh
  Ctrl+C      Quit
  ?           Show help
  ESC         Back/Cancel

Key Bindings (Keys view):
  d           Delete key
  e           Edit key
  t           Set TTL
  Enter       View key details
  r           Refresh keys
  
Key Bindings (Monitor view):
  s           Start/stop monitoring
  c           Clear screen
  r           Refresh
  
Key Bindings (CLI view):
  Enter       Execute command
  ↑/↓         Command history
  Ctrl+L      Clear screen

Examples:
  redis-cli-dashboard                                    # Connect to localhost:6379
  redis-cli-dashboard -host redis.example.com -port 6380  # Connect to remote server
  redis-cli-dashboard -password mypassword -db 1         # Connect with auth and DB selection

For more information, visit: https://github.com/username/redis-cli-dashboard
`)
}
