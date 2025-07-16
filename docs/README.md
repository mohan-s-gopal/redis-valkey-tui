
###  redis-cli-dashboard

> A powerful TUI client for Redis/Valkey — inspired by `k9s`.

---

### ✨ **Core Features** (Implemented)

* ✅ **Advanced TUI interface** with multiple modes
* ✅ **Multi-mode Navigation**: Key Browser, Monitoring, CLI, Analytics, Connections
* ✅ **Enhanced Key Management**:
  * View all keys with type awareness and TTL information
  * Filter keys by pattern (`user:*`, `session:*`, etc.)
  * Real-time search functionality
  * Key editing and value modification
  * TTL management (set, remove, view expiration)
  * Memory usage display per key
* ✅ **Complete Value Support**:
  * `string` (with JSON formatting)
  * `list` (with indices)
  * `set` (sorted display)
  * `hash` (field-value pairs)
  * `zset` (with scores)
  * `stream` (basic support)
* ✅ **Performance Monitoring**:
  * Real-time server metrics
  * Memory usage tracking
  * Connection monitoring
  * Operations per second
  * Cache hit/miss ratio
* ✅ **Built-in CLI**:
  * Redis command execution
  * Command history
  * Auto-completion support
  * Scrollable output
* ✅ **Advanced Features**:
  * Connection management
  * Multiple database support
  * Configuration file + CLI overrides
  * Error handling and recovery
  * Modal dialogs for confirmations

---

### 🧩 **Advanced Features** (Future Enhancements)

* 🔐 Enhanced Redis AUTH with ACL support
* 🧰 Redis Cluster mode support
* 🖥 TLS/SSL connections
* 💾 Export/Import functionality (JSON, CSV)
* � Advanced pattern matching and regex search
* 📈 Historical metrics and graphing
* 🪄 Multiple visual themes
* � Real-time key monitoring
* � Query builder for complex operations
* 🧪 Redis Stack modules support (RedisJSON, RedisSearch)
* � Syntax highlighting for commands
* 🔗 SSH tunneling support

---

````markdown
# 🔧 redis-cli-dashboard

A fast and minimal **TUI (terminal UI)** client for [Valkey](https://valkey.io) and Redis, inspired by [`k9s`](https://github.com/derailed/k9s). Built with Go and `tview`, `redis-cli-dashboard` gives you a real-time interactive interface to explore keys and manage data across all Redis types.

Here's a project name and feature list for your Redis/Valkey TUI tool:

---
---

## 📦 Features

- 🚀 Terminal UI using [`tview`](https://github.com/rivo/tview)
- 🔍 View all keys from the connected Redis/Valkey server
- 📂 Inspect value contents by type:
  - `string`
  - `list`
  - `set`
  - `hash`
  - `zset`
- 🔄 Refresh key list (`r`)
- 🗑️ Delete selected key (`d`)
- ⛔ Quit the app (`q`)

---

## 🖥️ Screenshots

> Coming soon...

---

## 🛠 Installation

### 1. Clone the repo

```bash
git clone https://github.com/your-username/redis-cli-dashboard.git
cd redis-cli-dashboard
```

### 2. Build

Make sure you have [Go](https://golang.org/dl/) 1.21+ installed.

```bash
make build
# or
go build -o redis-cli-dashboard main.go config.go cli.go
```

### 3. Run

```bash
./redis-cli-dashboard
```

## 🚀 Usage

### Basic Usage
```bash
./redis-cli-dashboard                                    # Connect to localhost:6379
./redis-cli-dashboard -host redis.example.com -port 6380  # Connect to remote server
./redis-cli-dashboard -password mypassword -db 1         # Connect with auth and DB selection
```

### Configuration File
Create `~/.redis-cli-dashboard/config.json`:
```json
{
  "redis": {
    "host": "localhost",
    "port": 6379,
    "password": "",
    "db": 0
  },
  "ui": {
    "theme": "dark",
    "refresh_ms": 1000
  }
}
```

### Command Line Options
- `-host string`: Redis host (default: localhost)
- `-port int`: Redis port (default: 6379)
- `-password string`: Redis password
- `-db int`: Redis database number (default: 0)
- `-help`: Show help message

---

## 🎯 Keybindings

### Global Navigation
| Key         | Action                |
| ----------- | --------------------- |
| `ESC`       | Return to main menu   |
| `F1` or `1` | Switch to Key Browser |
| `F2` or `2` | Switch to Monitoring  |
| `F3` or `3` | Switch to CLI Mode    |
| `F4` or `4` | Switch to Analytics   |
| `F5` or `5` | Switch to Connections |
| `q`         | Quit (from main menu) |

### Key Browser Mode
| Key     | Action              |
| ------- | ------------------- |
| `r`     | Refresh key list    |
| `d`     | Delete selected key |
| `e`     | Edit selected key   |
| `t`     | Set/modify TTL      |
| `f`     | Focus filter input  |
| `s`     | Focus search input  |
| `↑/↓`   | Navigate keys       |
| `Enter` | View key details    |

### Monitoring Mode
| Key   | Action          |
| ----- | --------------- |
| `r`   | Refresh metrics |
| `c`   | Clear display   |
| `↑/↓` | Scroll metrics  |

### CLI Mode
| Key      | Action          |
| -------- | --------------- |
| `Enter`  | Execute command |
| `↑/↓`    | Command history |
| `Ctrl+C` | Clear output    |

### Connection Mode
| Key | Action            |
| --- | ----------------- |
| `n` | New connection    |
| `e` | Edit connection   |
| `d` | Delete connection |

---

## 🔮 Roadmap

* [ ] Edit/add key values
* [ ] TTL display and update
* [ ] Support for cluster mode
* [ ] Authentication support
* [ ] Connect via TLS/SSL
* [ ] Pattern filter (`SCAN`, `KEYS prefix:*`)
* [ ] Export to JSON or YAML

---

## 📄 License

MIT License

---

## 🤝 Contributing

PRs welcome! Please open an issue or submit a pull request with improvements or feature additions.

---

## 🙌 Acknowledgments

* [Valkey](https://valkey.io)
* [Redis](https://redis.io)
* [tview](https://github.com/rivo/tview)
* [go-redis](https://github.com/go-redis/redis)

```

---

Let me know if you'd like a logo, Dockerfile, or GitHub Actions setup to go with this!
```
