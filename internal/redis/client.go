package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"redis-cli-dashboard/internal/config"

	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client with additional functionality
type Client struct {
	rdb *redis.Client
	ctx context.Context
}

// New creates a new Redis client
func New(cfg *config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

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

// GetKeys returns all keys matching the pattern
func (c *Client) GetKeys(pattern string) ([]string, error) {
	if pattern == "" {
		pattern = "*"
	}

	keys, err := c.rdb.Keys(c.ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}

	return keys, nil
}

// GetKeyInfo returns information about a key
func (c *Client) GetKeyInfo(key string) (*KeyInfo, error) {
	info := &KeyInfo{
		Key:  key,
		Name: key,
	}

	// Get key type
	keyType, err := c.rdb.Type(c.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get key type: %w", err)
	}
	info.Type = keyType

	// Get TTL
	ttl, err := c.rdb.TTL(c.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get TTL: %w", err)
	}
	info.TTL = ttl

	// Get memory usage (if supported)
	memUsage, err := c.rdb.MemoryUsage(c.ctx, key).Result()
	if err == nil {
		info.MemoryUsage = memUsage
		info.Size = memUsage // Set Size to match MemoryUsage
	}

	// Get encoding
	encoding, err := c.rdb.ObjectEncoding(c.ctx, key).Result()
	if err == nil {
		info.Encoding = encoding
	}

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
