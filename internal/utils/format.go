package utils

import (
	"fmt"
	"time"
)

// FormatUptime formats the uptime in a human-readable format
func FormatUptime(seconds int64) string {
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	if hours < 24 {
		return fmt.Sprintf("%dh", hours)
	}
	days := hours / 24
	return fmt.Sprintf("%dd", days)
}
