package ui

// Metrics holds Redis server metrics
type Metrics struct {
	Latency          int64
	MemoryUsed       uint64
	MemoryPeak       uint64
	OpsPerSec        int64
	ConnectedClients int64
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{}
}
