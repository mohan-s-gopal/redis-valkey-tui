package redis

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/mohan-s-gopal/redis-valkey-tui/internal/config"

	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client with additional functionality
type Client struct {
	rdb *redis.Client
	ctx context.Context
}

// New creates a new Redis client
func New(cfg *config.RedisConfig) (*Client, error) {
	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	// Configure TLS if enabled
	if cfg.TLS.Enabled {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		}

		// Load client certificate if provided
		if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFile, cfg.TLS.KeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load client certificate: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		// Load CA certificate if provided
		if cfg.TLS.CAFile != "" {
			caCert, err := ioutil.ReadFile(cfg.TLS.CAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA certificate: %w", err)
			}
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to parse CA certificate")
			}
			tlsConfig.RootCAs = caCertPool
		}

		opts.TLSConfig = tlsConfig
	}

	rdb := redis.NewClient(opts)

	ctx := context.Background()

	// Test connection
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{
		rdb: rdb,
		ctx: ctx,
	}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Type returns the type of a key
func (c *Client) Type(key string) (string, error) {
	return c.rdb.Type(c.ctx, key).Result()
}

// TTL returns the TTL of a key in seconds
func (c *Client) TTL(key string) (int64, error) {
	duration, err := c.rdb.TTL(c.ctx, key).Result()
	if err != nil {
		return -1, err
	}
	return int64(duration.Seconds()), nil
}

// MemoryUsage returns the memory usage of a key in bytes
func (c *Client) MemoryUsage(key string) (int64, error) {
	return c.rdb.MemoryUsage(c.ctx, key).Result()
}

// ObjectEncoding returns the internal encoding of a key
func (c *Client) ObjectEncoding(key string) (string, error) {
	return c.rdb.ObjectEncoding(c.ctx, key).Result()
}

// DBSize returns the number of keys in the current database
func (c *Client) DBSize() (int64, error) {
	return c.rdb.DBSize(c.ctx).Result()
}

// GetKeys returns all keys matching the pattern using SCAN for safety
func (c *Client) GetKeys(pattern string) ([]string, error) {
	if pattern == "" {
		pattern = "*"
	}

	var keys []string
	var cursor uint64
	
	// Use SCAN instead of KEYS for better performance and cluster compatibility
	for {
		var scanKeys []string
		var err error
		
		scanKeys, cursor, err = c.rdb.Scan(c.ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan keys: %w", err)
		}
		
		keys = append(keys, scanKeys...)
		
		// Break when cursor returns to 0 (full scan complete)
		if cursor == 0 {
			break
		}
		
		// Safety check to prevent infinite loops
		if len(keys) > 10000 {
			break
		}
	}

	return keys, nil
}

// GetKeyInfo returns information about a key
func (c *Client) GetKeyInfo(key string) (*KeyInfo, error) {
	info := &KeyInfo{
		Key:  key,
		Name: key,
		Type: "unknown", // Default fallback
		TTL:  -1,        // Default to no expiry
	}

	// Get key type - this is critical, so return error if it fails
	keyType, err := c.rdb.Type(c.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get key type for %s: %w", key, err)
	}
	info.Type = keyType

	// Get TTL - don't fail if this doesn't work
	ttl, err := c.rdb.TTL(c.ctx, key).Result()
	if err == nil {
		info.TTL = ttl
	}

	// Get memory usage (if supported) - don't fail if this doesn't work
	memUsage, err := c.rdb.MemoryUsage(c.ctx, key).Result()
	if err == nil {
		info.MemoryUsage = memUsage
		info.Size = memUsage // Set Size to match MemoryUsage
	} else {
		// Try to get approximate size based on key type
		info.Size = c.getApproximateSize(key, keyType)
	}

	// Skip encoding for now since we removed it from UI
	// and it can cause issues in some Redis configurations

	return info, nil
}

// GetValue returns the value of a key
func (c *Client) GetValue(key string) (string, error) {
	keyType, err := c.rdb.Type(c.ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get key type: %w", err)
	}

	switch keyType {
	case "string":
		return c.rdb.Get(c.ctx, key).Result()
	case "list":
		values, err := c.rdb.LRange(c.ctx, key, 0, -1).Result()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("[%s]", strings.Join(values, ", ")), nil
	case "set":
		values, err := c.rdb.SMembers(c.ctx, key).Result()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("{%s}", strings.Join(values, ", ")), nil
	case "hash":
		values, err := c.rdb.HGetAll(c.ctx, key).Result()
		if err != nil {
			return "", err
		}
		var pairs []string
		for k, v := range values {
			pairs = append(pairs, fmt.Sprintf("%s: %s", k, v))
		}
		return fmt.Sprintf("{%s}", strings.Join(pairs, ", ")), nil
	case "zset":
		values, err := c.rdb.ZRangeWithScores(c.ctx, key, 0, -1).Result()
		if err != nil {
			return "", err
		}
		var pairs []string
		for _, z := range values {
			pairs = append(pairs, fmt.Sprintf("%v: %.2f", z.Member, z.Score))
		}
		return fmt.Sprintf("[%s]", strings.Join(pairs, ", ")), nil
	default:
		return "", fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// SetValue sets the value of a key
func (c *Client) SetValue(key, value string) error {
	return c.rdb.Set(c.ctx, key, value, 0).Err()
}

// DeleteKey deletes a key
func (c *Client) DeleteKey(key string) error {
	return c.rdb.Del(c.ctx, key).Err()
}

// SetTTL sets the TTL for a key
func (c *Client) SetTTL(key string, ttl time.Duration) error {
	return c.rdb.Expire(c.ctx, key, ttl).Err()
}

// GetInfo returns server info
func (c *Client) GetInfo() (map[string]string, error) {
	info, err := c.rdb.Info(c.ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	result := make(map[string]string)
	lines := strings.Split(info, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result, nil
}

// ExecuteCommand executes a Redis command
func (c *Client) ExecuteCommand(cmd string, args ...interface{}) (interface{}, error) {
	cmdArgs := append([]interface{}{cmd}, args...)
	return c.rdb.Do(c.ctx, cmdArgs...).Result()
}

// Info returns Redis INFO command output
func (c *Client) Info() (map[string]interface{}, error) {
	info, err := c.rdb.Info(c.ctx).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	lines := strings.Split(info, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := strings.TrimSpace(parts[1])

			// Try to convert to integer
			if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
				result[key] = intVal
				continue
			}

			// Try to convert to float
			if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
				result[key] = floatVal
				continue
			}

			// Keep as string
			result[key] = value
		}
	}

	return result, nil
}

// GetMetrics returns Redis metrics
func (c *Client) GetMetrics() (*Metrics, error) {
	info, err := c.GetInfo()
	if err != nil {
		return nil, err
	}

	metrics := &Metrics{}

	if val, ok := info["connected_clients"]; ok {
		metrics.ConnectedClients, _ = strconv.ParseInt(val, 10, 64)
	}
	if val, ok := info["used_memory"]; ok {
		if parsed, err := strconv.ParseUint(val, 10, 64); err == nil {
			metrics.UsedMemory = parsed
		}
	}
	if val, ok := info["used_memory_rss"]; ok {
		if parsed, err := strconv.ParseUint(val, 10, 64); err == nil {
			metrics.UsedMemoryRss = parsed
		}
	}
	if val, ok := info["total_commands_processed"]; ok {
		metrics.TotalCommandsProcessed, _ = strconv.ParseInt(val, 10, 64)
	}
	if val, ok := info["keyspace_hits"]; ok {
		metrics.KeyspaceHits, _ = strconv.ParseInt(val, 10, 64)
	}
	if val, ok := info["keyspace_misses"]; ok {
		metrics.KeyspaceMisses, _ = strconv.ParseInt(val, 10, 64)
	}
	if val, ok := info["instantaneous_ops_per_sec"]; ok {
		metrics.InstantaneousOpsPerSec, _ = strconv.ParseInt(val, 10, 64)
	}
	if val, ok := info["uptime_in_seconds"]; ok {
		metrics.UptimeInSeconds, _ = strconv.ParseInt(val, 10, 64)
	}

	return metrics, nil
}

// KeyInfo holds information about a Redis key
type KeyInfo struct {
	Key         string
	Name        string
	Type        string
	TTL         time.Duration
	Size        int64
	Encoding    string
	MemoryUsage int64
}

// Metrics holds Redis server metrics
type Metrics struct {
	ConnectedClients       int64
	UsedMemory             uint64
	UsedMemoryRss          uint64
	TotalCommandsProcessed int64
	KeyspaceHits           int64
	KeyspaceMisses         int64
	InstantaneousOpsPerSec int64
	UptimeInSeconds        int64
}

// getApproximateSize tries to get an approximate size for a key when MemoryUsage fails
func (c *Client) getApproximateSize(key, keyType string) int64 {
	switch keyType {
	case "string":
		if val, err := c.rdb.Get(c.ctx, key).Result(); err == nil {
			return int64(len(val))
		}
	case "list":
		if length, err := c.rdb.LLen(c.ctx, key).Result(); err == nil {
			return length * 50 // Rough estimate
		}
	case "set":
		if length, err := c.rdb.SCard(c.ctx, key).Result(); err == nil {
			return length * 50 // Rough estimate
		}
	case "hash":
		if length, err := c.rdb.HLen(c.ctx, key).Result(); err == nil {
			return length * 100 // Rough estimate
		}
	case "zset":
		if length, err := c.rdb.ZCard(c.ctx, key).Result(); err == nil {
			return length * 100 // Rough estimate
		}
	}
	return 0 // Unknown
}

// ClusterNodes returns cluster nodes information
func (c *Client) ClusterNodes() (string, error) {
	return c.rdb.ClusterNodes(c.ctx).Result()
}

// CommandStat represents statistics for a single Redis command
type CommandStat struct {
	Command        string
	Calls          int64
	TotalDuration  float64 // in milliseconds
	DurationPerCall float64 // in milliseconds
	RejectedCalls  int64
	FailedCalls    int64
}

// GetCommandStats returns command statistics from Redis INFO commandstats
func (c *Client) GetCommandStats() ([]CommandStat, error) {
	info, err := c.rdb.Info(c.ctx, "commandstats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get command stats: %w", err)
	}

	var stats []CommandStat
	lines := strings.Split(info, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "cmdstat_") {
			continue
		}

		// Parse line like: cmdstat_info:calls=90,usec=100900,usec_per_call=1.12,rejected_calls=0,failed_calls=0
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		command := strings.TrimPrefix(parts[0], "cmdstat_")
		values := parts[1]

		stat := CommandStat{Command: command}

		// Parse key=value pairs
		pairs := strings.Split(values, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) != 2 {
				continue
			}

			key, value := kv[0], kv[1]
			
			switch key {
			case "calls":
				if val, err := strconv.ParseInt(value, 10, 64); err == nil {
					stat.Calls = val
				}
			case "usec":
				if val, err := strconv.ParseFloat(value, 64); err == nil {
					stat.TotalDuration = val / 1000.0 // Convert microseconds to milliseconds
				}
			case "usec_per_call":
				if val, err := strconv.ParseFloat(value, 64); err == nil {
					stat.DurationPerCall = val / 1000.0 // Convert microseconds to milliseconds
				}
			case "rejected_calls":
				if val, err := strconv.ParseInt(value, 10, 64); err == nil {
					stat.RejectedCalls = val
				}
			case "failed_calls":
				if val, err := strconv.ParseInt(value, 10, 64); err == nil {
					stat.FailedCalls = val
				}
			}
		}

		stats = append(stats, stat)
	}

	return stats, nil
}

// ClientInfo represents information about a connected client
type ClientInfo struct {
	ID            string
	Address       string
	Age           int64  // Connection age in seconds
	Idle          int64  // Idle time in seconds
	LastCommand   string
	DB            int
	Name          string
	TotalDuration float64 // Calculated from age
}

// GetClientList returns information about connected clients
func (c *Client) GetClientList() ([]ClientInfo, error) {
	result, err := c.rdb.ClientList(c.ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get client list: %w", err)
	}

	var clients []ClientInfo
	lines := strings.Split(result, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		client := ClientInfo{}
		
		// Parse client info line with key=value pairs
		pairs := strings.Split(line, " ")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) != 2 {
				continue
			}

			key, value := kv[0], kv[1]
			
			switch key {
			case "id":
				client.ID = value
			case "addr":
				client.Address = value
			case "age":
				if val, err := strconv.ParseInt(value, 10, 64); err == nil {
					client.Age = val
					// Calculate total duration in minutes from age
					client.TotalDuration = float64(val) / 60.0
				}
			case "idle":
				if val, err := strconv.ParseInt(value, 10, 64); err == nil {
					client.Idle = val
				}
			case "cmd":
				client.LastCommand = value
			case "db":
				if val, err := strconv.Atoi(value); err == nil {
					client.DB = val
				}
			case "name":
				client.Name = value
			}
		}

		clients = append(clients, client)
	}

	return clients, nil
}
