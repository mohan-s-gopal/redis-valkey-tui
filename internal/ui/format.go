package ui

import "github.com/dustin/go-humanize"

// formatBytes formats a byte count into a human-readable string
func formatBytes(bytes uint64) string {
	return humanize.Bytes(bytes)
}
