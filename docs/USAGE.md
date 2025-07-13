# valkys Usage Guide

## Quick Start

1. **Build the application:**
   ```bash
   make build
   ```

2. **Run with default settings:**
   ```bash
   ./valkys
   ```

3. **Run with custom Redis connection:**
   ```bash
   ./valkys -host redis.example.com -port 6380 -password mypassword
   ```

## Command Line Options

- `-host string`: Redis host (default: localhost)
- `-port int`: Redis port (default: 6379)
- `-password string`: Redis password
- `-db int`: Redis database number (default: 0)
- `-help`: Show help message

## Configuration File

valkys will look for a configuration file at `~/.valkys/config.json`. Example:

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

Command line options override config file values.

## Application Modes

valkys features multiple modes for different tasks:

### 1. Key Browser Mode (Default)
- **Purpose**: Browse, search, and manage Redis keys
- **Features**:
  - Filter keys by pattern
  - Real-time search
  - View key values with syntax highlighting
  - Edit string values
  - Set/remove TTL
  - Delete keys with confirmation
  - Memory usage display

### 2. Monitoring Mode
- **Purpose**: Monitor Redis server performance
- **Features**:
  - Real-time server metrics
  - Memory usage tracking
  - Operations per second
  - Cache hit/miss ratio
  - Connection statistics
  - Server uptime

### 3. CLI Mode
- **Purpose**: Execute Redis commands directly
- **Features**:
  - Full Redis command support
  - Command history
  - Scrollable output
  - Auto-completion
  - Error handling

### 4. Analytics Mode
- **Purpose**: Analyze Redis data and performance
- **Features**:
  - Advanced metrics
  - Data analysis tools
  - Performance insights

### 5. Connections Mode
- **Purpose**: Manage multiple Redis connections
- **Features**:
  - Switch between connections
  - Add/edit/remove connections
  - Connection status monitoring

## Navigation

### Main Menu
When you start valkys, you'll see the main menu with 5 options:
- Use `↑/↓` to navigate
- Press `Enter` to select
- Press `q` to quit

### Quick Mode Switching
- `F1` or `1`: Key Browser
- `F2` or `2`: Monitoring  
- `F3` or `3`: CLI
- `F4` or `4`: Analytics
- `F5` or `5`: Connections
- `ESC`: Return to main menu

## Key Browser Mode

### Filtering and Searching
1. **Pattern Filtering**: Press `f` to focus the filter input
   - Enter patterns like `user:*`, `session:*`, `cache:*`
   - Press `Enter` to apply the filter
   
2. **Real-time Search**: Press `s` to focus the search input
   - Type to search keys in real-time
   - Search is case-insensitive
   
### Key Operations
1. **View Key**: Select a key and press `Enter`
   - Shows value formatted by type
   - JSON strings are automatically formatted
   - TTL information displayed

2. **Edit Key**: Press `e` on a selected key
   - Currently supports string keys only
   - Can change both key name and value

3. **Set TTL**: Press `t` on a selected key
   - Set expiration time in seconds
   - Remove TTL to make key persistent
   - View current TTL status

4. **Delete Key**: Press `d` on a selected key
   - Confirmation dialog appears
   - Permanently removes the key

## Monitoring Mode

### Metrics Display
- **Connection Info**: Connected clients
- **Memory Usage**: Used memory and RSS
- **Performance**: Operations per second, total commands
- **Cache Stats**: Hit/miss ratio, keyspace statistics
- **Server Info**: Uptime, connection details

### Refresh
- Press `r` to refresh metrics
- Press `c` to clear display
- Use `↑/↓` to scroll through metrics

## CLI Mode

### Supported Commands
- `PING`: Test connection
- `INFO`: Get server information
- `KEYS pattern`: List keys matching pattern
- `GET key`: Get key value
- `SET key value`: Set key value
- `DEL key [key ...]`: Delete keys
- `EXISTS key [key ...]`: Check key existence
- `TYPE key`: Get key type
- `TTL key`: Get key TTL
- `EXPIRE key seconds`: Set key expiration

### Command History
- Use `↑/↓` to navigate command history
- Press `Enter` to execute commands
- Press `Ctrl+C` to clear output

## Data Type Support

### String
- Raw text display
- Automatic JSON formatting
- Inline editing support

### List
- Indexed display: `[0] value1`, `[1] value2`
- Shows list length
- Supports all Redis list operations

### Set
- Alphabetically sorted display
- Shows member count
- Supports all Redis set operations

### Hash
- Field-value pairs display
- Shows field count
- Supports all Redis hash operations

### Sorted Set (ZSet)
- Members with scores: `member (score: 1.5)`
- Sorted by score
- Shows member count

### Stream
- Basic stream entry display
- Shows entry ID and field-value pairs
- Entry count display

## Keybindings

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

## Examples

### Basic Usage
```bash
# Connect to local Redis
./valkys

# Connect to remote Redis with authentication
./valkys -host redis.example.com -port 6380 -password secret123

# Connect to specific database
./valkys -db 2
```

### Using Configuration File
Create `~/.valkys/config.json`:
```json
{
  "redis": {
    "host": "redis.example.com",
    "port": 6380,
    "password": "secret123",
    "db": 1
  }
}
```

Then simply run:
```bash
./valkys
```

### Overriding Configuration
```bash
# Use config file but override the database
./valkys -db 3

# Use config file but override the host
./valkys -host localhost
```

## Troubleshooting

### Connection Issues
- Ensure Redis/Valkey server is running
- Check host and port settings
- Verify authentication credentials
- Check network connectivity

### Performance
- For large datasets, consider using Redis SCAN instead of KEYS
- The application refreshes the key list when 'r' is pressed
- Large values may take time to display

### Building Issues
- Ensure Go 1.21+ is installed
- Run `go mod tidy` to resolve dependencies
- Check that all required packages are available
