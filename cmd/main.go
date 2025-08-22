package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mohan-s-gopal/redis-valkey-tui/internal/config"
	"github.com/mohan-s-gopal/redis-valkey-tui/internal/logger"
	"github.com/mohan-s-gopal/redis-valkey-tui/internal/ui"
)

// Version information
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// Main is the main entry point
func Main() {
	var (
		host     = flag.String("host", "", "Redis host")
		port     = flag.Int("port", 0, "Redis port")
		password = flag.String("password", "", "Redis password")
		db       = flag.Int("db", -1, "Redis database number")
		verbose  = flag.Int("v", 0, "Verbosity level (0=ERROR, 1=WARN, 2=INFO, 3=DEBUG, 4=TRACE)")
		console  = flag.Bool("console", false, "Enable console logging (logs will appear in stderr)")
		help     = flag.Bool("help", false, "Show help")
		version  = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *version {
		showVersion()
		return
	}

	if *help {
		showHelp()
		return
	}

	// Initialize logger with verbosity level
	var logLevel logger.LogLevel
	switch *verbose {
	case 0:
		logLevel = logger.ERROR
	case 1:
		logLevel = logger.WARN
	case 2:
		logLevel = logger.INFO
	case 3:
		logLevel = logger.DEBUG
	case 4:
		logLevel = logger.TRACE
	default:
		logLevel = logger.INFO
	}

	if err := logger.InitWithLevel(logLevel, *console); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("Starting redis-valkey-tui...")
	logger.Debugf("Verbosity level: %d, Console output: %t", *verbose, *console)

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

func showVersion() {
	fmt.Printf("redis-valkey-tui version %s\n", Version)
	fmt.Printf("Build time: %s\n", BuildTime)
	fmt.Printf("Git commit: %s\n", GitCommit)
}

func showHelp() {
	fmt.Print(`redis-valkey-tui - A k9s-inspired TUI client for Redis/Valkey

Usage: redis-valkey-tui [options]

Options:
  -host string
        Redis host (default: localhost)
  -port int
        Redis port (default: 6379)
  -password string
        Redis password
  -db int
        Redis database number (default: 0)
  -v int
        Verbosity level: 0=ERROR, 1=WARN, 2=INFO, 3=DEBUG, 4=TRACE (default: 2)
  -console
        Enable console logging (logs will also appear in stderr)
  -help
        Show this help
  -version
        Show version information

Logging:
  By default, logs are written to ~/.redis-valkey-tui/logs/app.log
  Use -v 3 or -v 4 for detailed debugging information
  Use -console to see logs in terminal while app is running

Navigation:
  1-6         Switch to different views (1=Keys, 2=Info, 3=Monitor, 4=CLI, 5=Config, 6=Help)
  
  /           Filter/search
  Ctrl+R      Refresh
  Ctrl+C      Quit
  ?           Show help
  ESC         Back/Cancel

Key Bindings (Keys view):
  c           Execute command
  /           Filter keys
  r           Refresh keys
  Enter       View key details
  
Key Bindings (Monitor view):
  s           Start/stop monitoring
  c           Clear screen
  r           Refresh
  
Key Bindings (CLI view):
  Enter       Execute command
  ↑/↓         Command history
  Ctrl+L      Clear screen

Examples:
  redis-valkey-tui                                      # Connect to localhost:6379 with INFO logging
  redis-valkey-tui -v 4 -console                       # Run with TRACE logging to console
  redis-valkey-tui -host redis.example.com -port 6380  # Connect to remote server
  redis-valkey-tui -password mypassword -db 1 -v 3     # Connect with auth, DB selection, and DEBUG logging
  redis-valkey-tui --version                           # Show version information

For debugging issues, use: redis-valkey-tui -v 4 -console

For more information, visit: https://github.com/mohan-s-gopal/redis-valkey-tui
`)
}
