// package ids (or keep in conversation package)
package ids

import (
	"sync"
	"time"
)

// Snowflake:  41 bits timestamp | 10 bits node | 12 bits seq
type Snowflake struct {
	mu      sync.Mutex
	epochMs int64
	node    int64 // 0..1023
	seq     int64 // 0..4095
	lastTs  int64
}

func NewSnowflake(node int64) *Snowflake {
	if node < 0 || node > 1023 {
		panic("snowflake node must be in [0,1023]")
	}
	// pick your own fixed epoch (smaller numbers -> more future years)
	const customEpoch = int64(1704067200000) // 2024-01-01 UTC in ms

	return &Snowflake{epochMs: customEpoch, node: node}
}

func (s *Snowflake) Next() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli() - s.epochMs
	if now == s.lastTs {
		s.seq = (s.seq + 1) & 0xFFF // 12-bit sequence
		if s.seq == 0 {
			// sequence exhausted in this millisecond; spin to next ms
			for now <= s.lastTs {
				now = time.Now().UnixMilli() - s.epochMs
			}
		}
	} else {
		s.seq = 0
	}

	s.lastTs = now

	return (now << 22) | (s.node << 12) | s.seq
}
