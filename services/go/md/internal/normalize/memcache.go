package normalize

import (
	"sync"
	"time"
)

type memCache struct {
	mu sync.RWMutex
	m  map[string]memEntry
}

type memEntry struct {
	row   instrumentRow
	until time.Time
}

func newMemCache() *memCache {
	return &memCache{m: make(map[string]memEntry)}
}

func (c *memCache) get(id string) (instrumentRow, bool) {
	c.mu.RLock()
	e, ok := c.m[id]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.until) {
		return instrumentRow{}, false
	}
	return e.row, true
}

func (c *memCache) set(id string, row instrumentRow, ttl time.Duration) {
	c.mu.Lock()
	c.m[id] = memEntry{row: row, until: time.Now().Add(ttl)}
	c.mu.Unlock()
}
