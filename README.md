# redis-cli-dashboard - A k9s-inspired TUI Redis/Valkey Client

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A powerful Terminal User Interface (TUI) client for Redis and Valkey, inspired by k9s navigation patterns and RedisInsight features.

## 🌟 Features

### 🎯 Multi-View Navigation
- **Keys View**: Browse, filter, and manage Redis keys with real-time search
- **Info View**: Display comprehensive server information and metrics
- **Monitor View**: Real-time performance monitoring and metrics
- **CLI View**: Built-in Redis command line interface with history
- **Config View**: Configuration management and settings

### 🔧 Advanced Capabilities
- **k9s-inspired Navigation**: Intuitive command-driven interface
- **Real-time Filtering**: Pattern-based key filtering and search
- **TTL Management**: View and modify key expiration times
- **Memory Analytics**: Track memory usage per key and overall
- **JSON Formatting**: Automatic JSON detection and pretty printing
- **Command History**: Navigate through previously executed commands
- **Multi-database Support**: Switch between Redis databases

### 🎨 User Experience
- **Responsive TUI**: Built with tview for a smooth terminal experience
- **Keyboard Shortcuts**: Efficient navigation with keyboard shortcuts
- **Status Indicators**: Real-time status updates and feedback
- **Help System**: Built-in help with command reference
- **Error Handling**: Graceful error handling with user feedback

## 📁 Project Structure

```
redis-cli-dashboard/
├── cmd/                    # Application entry point
│   └── main.go            # Main command package
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   │   └── config.go     # Config types and functions
│   ├── redis/            # Redis client wrapper
│   │   └── client.go     # Redis operations and metrics
│   ├── ui/               # User interface components
│   │   ├── app.go        # Main application logic
│   │   ├── keys_view.go  # Keys browser view
│   │   ├── info_view.go  # Server info view
│   │   ├── monitor_view.go # Monitoring view
│   │   ├── cli_view.go   # CLI interface view
│   │   └── config_view.go # Configuration view
│   └── views/            # Additional view components
├── pkg/                  # Public packages (if any)
│   └── utils/           # Utility functions
├── docs/                 # Documentation
│   ├── README.md        # Main documentation
│   ├── USAGE.md         # Usage guide
│   ├── Features.md      # Feature documentation
│   └── *.md            # Other documentation files
├── scripts/             # Build and setup scripts
│   ├── setup.sh        # Development setup
│   └── demo.sh         # Feature demonstration
├── .github/             # GitHub workflows
│   └── workflows/
│       └── ci.yml      # CI/CD pipeline
├── main.go             # Application entry point
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
├── Makefile           # Build automation
├── Dockerfile         # Container definition
└── LICENSE            # MIT license
```

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or later
- Redis or Valkey server

### Installation

#### Option 1: Quick Setup
```bash
# Clone the repository
git clone https://github.com/username/redis-cli-dashboard.git
cd redis-cli-dashboard

# Run setup script
./scripts/setup.sh
```

#### Option 2: Manual Build
```bash
# Install dependencies
go mod download

# Build the application
make build

# Run the application
./build/redis-cli-dashboard
```

#### Option 3: Using Docker
```bash
# Build Docker image
docker build -t redis-cli-dashboard .

# Run with Docker
docker run -it --network host redis-cli-dashboard
```

### Basic Usage

```bash
# Connect to local Redis (default: localhost:6379)
redis-cli-dashboard

# Connect to remote server
redis-cli-dashboard -host redis.example.com -port 6380

# Connect with authentication
redis-cli-dashboard -password mypassword -db 1

# Show help
redis-cli-dashboard -help
```

## 🎮 Navigation

### Command Mode
Press `:` to enter command mode, then type:
- `:keys` - Switch to Keys view
- `:info` - Switch to Info view  
- `:monitor` - Switch to Monitor view
- `:cli` - Switch to CLI view
- `:config` - Switch to Config view
- `:quit` or `:q` - Quit application

### Global Key Bindings
- `Ctrl+C` - Quit application
- `Ctrl+R` - Refresh current view
- `?` - Show help
- `ESC` - Back/Cancel

### View-Specific Controls

#### Keys View
- `d` - Delete selected key
- `e` - Edit selected key
- `t` - Set TTL for selected key
- `/` - Filter keys
- `r` - Refresh keys
- `Enter` - View key details

#### Monitor View
- `s` - Start/stop monitoring
- `c` - Clear screen
- `r` - Refresh metrics

#### CLI View
- `Enter` - Execute command
- `↑/↓` - Navigate command history
- `Ctrl+L` - Clear screen

## ⚙️ Configuration

### Configuration File
redis-cli-dashboard looks for configuration at `~/.redis-cli-dashboard/config.json`:

```json
{
  "redis": {
    "host": "localhost",
    "port": 6379,
    "password": "",
    "db": 0,
    "timeout": 5000,
    "pool_size": 10
  },
  "ui": {
    "theme": "default",
    "refresh_interval": 1000,
    "max_keys": 1000,
    "show_memory": true,
    "show_ttl": true
  }
}
```

### Command Line Options
Command line options override configuration file values:

```bash
redis-cli-dashboard -host localhost -port 6379 -password secret -db 1
```

## 🔧 Development

### Building

```bash
# Development build
make dev

# Production build
make build

# Build for all platforms
make build-all

# Run tests
make test

# Format code
make fmt

# Lint code
make lint
```

### Project Layout

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout):

- `cmd/` - Main applications for this project
- `internal/` - Private application and library code
- `pkg/` - Library code that's safe to use by external applications
- `docs/` - Documentation files
- `scripts/` - Scripts for building, installation, analysis, etc.

### Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   cmd/main.go   │    │  internal/ui/   │    │ internal/redis/ │
│                 │───▶│     app.go      │───▶│   client.go     │
│   Entry Point   │    │  View Manager   │    │ Redis Operations│
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │ internal/config/│
                    │   config.go     │
                    │  Configuration  │
                    └─────────────────┘
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by [k9s](https://github.com/derailed/k9s) for navigation patterns
- Inspired by [RedisInsight](https://github.com/RedisInsight/RedisInsight) for features
- Built with [tview](https://github.com/rivo/tview) for the TUI framework
- Uses [go-redis](https://github.com/redis/go-redis) for Redis connectivity

---

**🔴 Made with ❤️ for the Redis/Valkey community**
