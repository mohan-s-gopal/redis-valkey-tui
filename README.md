# redis-cli-dashboard

A powerful, k9s-inspired Terminal User Interface (TUI) Redis/Valkey client built with Go. Provides an intuitive, keyboard-driven interface for Redis database management with real-time monitoring, advanced key management, and built-in CLI functionality.

![Redis CLI Dashboard](docs/screenshot.png)

## 🌟 Features

### 🎯 **Multi-View Navigation**
- **6 Specialized Views**: Keys, Info, Monitor, CLI, Config, and Help
- **Instant View Switching**: Number keys (1-6) for immediate navigation
- **Home Navigation**: ESC key returns to main screen (Keys view)
- **Clean Interface**: No blocking or hanging during view transitions

### 🔧 **Advanced Key Management**
- **Real-time Key Browser**: Browse Redis keys with type indicators and metadata
- **Smart Filtering**: Live search and pattern matching with `/` key
- **Key Operations**: View, edit, delete keys with TTL management
- **Type Support**: Strings, Lists, Sets, Hashes, Sorted Sets with proper formatting
- **Memory Usage**: Display key sizes and memory consumption

### 📊 **Real-time Monitoring**
- **Live Metrics**: Server info, memory usage, connected clients
- **Performance Tracking**: Operations per second, command statistics
- **Connection Status**: Real-time connection health monitoring
- **Header Display**: Always-visible server status and database info

### 💻 **Integrated CLI**
- **Full Redis CLI**: Execute any Redis command directly in the interface
- **Command History**: Navigate previous commands with arrow keys
- **Syntax Support**: Proper command formatting and response display
- **Quick Access**: Switch between GUI and CLI modes instantly

### ⚙️ **Configuration Management**
- **Live Config View**: Display current Redis and application settings
- **Connection Management**: Host, port, database, authentication settings
- **UI Preferences**: Theme, refresh intervals, display options
- **Save/Reset**: Configuration persistence and defaults

### 🎨 **User Experience**
- **Responsive TUI**: Built with tview for smooth terminal experience
- **Keyboard Shortcuts**: Efficient navigation with intuitive key bindings
- **Status Indicators**: Real-time feedback and operation status
- **Help System**: Built-in help with comprehensive command reference
- **Error Handling**: Graceful error handling with user-friendly messages

## 🚀 Quick Start

### Installation

```bash
git clone https://github.com/mohan-s-gopal/redis-cli-dashboard.git
cd redis-cli-dashboard
go build -o redis-cli-dashboard ./cmd
```

### Running

```bash
# Start with default settings (localhost:6379)
./redis-cli-dashboard

# Connect to specific Redis instance
./redis-cli-dashboard -host redis.example.com -port 6380 -db 1

# Enable debug logging
./redis-cli-dashboard -v 2
```

### Configuration

Create `~/.redis-cli-dashboard/config.yaml`:

```yaml
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  timeout: 5000
  pool_size: 10

ui:
  theme: "default"
  refresh_interval: 1000
  max_keys: 1000
  show_memory: true
  show_ttl: true
```

## 🎮 Navigation & Controls

### Global Navigation
| Key | Action |
|-----|--------|
| `1` | Switch to Keys view (main screen) |
| `2` | Switch to Info view |
| `3` | Switch to Monitor view |
| `4` | Switch to CLI view |
| `5` | Switch to Config view |
| `6` | Switch to Help view |
| `ESC` | Return to main screen (Keys view) |
| `?` | Show help modal |
| `Ctrl+C` | Quit application |
| `Ctrl+R` | Refresh current view |

### Keys View
| Key | Action |
|-----|--------|
| `↑/↓` | Navigate keys |
| `Enter` | View key details |
| `/` | Focus filter input |
| `c` | Focus command input |
| `r` | Refresh key list |
| `d` | Delete selected key |
| `e` | Edit selected key |
| `t` | Set/modify TTL |

### CLI View
| Key | Action |
|-----|--------|
| `Enter` | Execute command |
| `↑/↓` | Navigate command history |
| `Ctrl+L` | Clear screen |
| `Tab` | Command completion (if available) |

### Monitor View
| Key | Action |
|-----|--------|
| `s` | Start/stop monitoring |
| `c` | Clear screen |
| `r` | Refresh metrics |

## 🏗️ Architecture

### Project Structure
```
redis-cli-dashboard/
├── cmd/                    # Application entry point
├── internal/
│   ├── config/            # Configuration management
│   ├── logger/            # Logging system
│   ├── redis/             # Redis client wrapper
│   ├── ui/                # TUI components
│   │   ├── app.go         # Main application
│   │   ├── keys_view.go   # Key browser
│   │   ├── cli_view.go    # CLI interface
│   │   ├── monitor_view.go # Monitoring
│   │   └── ...            # Other views
│   └── utils/             # Utility functions
├── docs/                  # Documentation
└── README.md
```

### Key Components
- **TUI Framework**: [tview](https://github.com/rivo/tview) for terminal interface
- **Redis Client**: [go-redis](https://github.com/redis/go-redis) for Redis connectivity
- **Configuration**: YAML-based configuration with defaults
- **Logging**: Structured logging with configurable verbosity
- **Testing**: Comprehensive test coverage for core functionality

## 🔧 Development

### Prerequisites
- Go 1.19 or later
- Redis/Valkey server for testing
- Terminal with color support

### Building from Source
```bash
git clone https://github.com/mohan-s-gopal/redis-cli-dashboard.git
cd redis-cli-dashboard
go mod tidy
go build -o redis-cli-dashboard ./cmd
```

### Running Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -v -cover ./...

# Run specific test
go test -v ./internal/redis
```

### Development Workflow
1. Make changes to source code
2. Run tests to ensure functionality
3. Build and test the application
4. Submit pull request with description

## 📋 Command Line Options

```
Usage: redis-cli-dashboard [OPTIONS]

Connection Options:
  -host string        Redis host (default "localhost")
  -port int          Redis port (default 6379)
  -password string   Redis password
  -db int           Redis database number (default 0)

Application Options:
  -config string     Config file path (default "~/.redis-cli-dashboard/config.yaml")
  -v int            Verbosity level 0-4 (default 0)
  -version          Show version information

Examples:
  redis-cli-dashboard                                    # Connect to localhost:6379
  redis-cli-dashboard -host prod.redis.com -port 6380   # Connect to remote Redis
  redis-cli-dashboard -db 2 -v 2                        # Use database 2 with debug logging
```

## 🔍 Troubleshooting

### Common Issues

**Connection Failed**
```bash
# Check Redis server status
redis-cli ping

# Test with specific host/port
redis-cli -h your-host -p your-port ping
```

**UI Not Displaying Correctly**
- Ensure terminal supports color and has sufficient size (minimum 80x24)
- Try different terminal emulators if issues persist
- Check terminal environment variables (`TERM`, `COLORTERM`)

**High Memory Usage**
- Reduce `max_keys` in configuration
- Lower `refresh_interval` to reduce update frequency
- Use filtering to limit displayed keys

### Debug Mode
```bash
# Enable detailed logging
./redis-cli-dashboard -v 4

# Check logs
tail -f ~/.redis-cli-dashboard/logs/app.log
```

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Areas for Contribution
- Additional Redis data type support
- Performance optimizations
- UI/UX improvements
- Documentation and examples
- Test coverage expansion
- Bug fixes and stability improvements

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by [k9s](https://github.com/derailed/k9s) for Kubernetes
- Built with [tview](https://github.com/rivo/tview) TUI framework
- Uses [go-redis](https://github.com/redis/go-redis) for Redis connectivity
- Community feedback and contributions

## 🔗 Links

- **Repository**: https://github.com/mohan-s-gopal/redis-cli-dashboard
- **Issues**: https://github.com/mohan-s-gopal/redis-cli-dashboard/issues
- **Documentation**: [docs/](docs/)
- **Releases**: https://github.com/mohan-s-gopal/redis-cli-dashboard/releases

---

**Made with ❤️ for the Redis community**
