package pokecache

import (
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

type Cache struct {
	entries map[string]cacheEntry
	mu      *sync.Mutex
}

type cacheEntry struct {
	createdAt time.Time
	data      []byte
}

func NewCache(interval time.Duration) *Cache {
	result := &Cache{
        entries: make(map[string]cacheEntry),
        mu: new(sync.Mutex),
    }
    log.Debug("Initiating cache reaping loop")
	go result.reapLoop(interval)
	return result
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{
		createdAt: time.Now(),
		data:      val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.entries[key]
	return val.data, ok
}

func (c *Cache) reapLoop(interval time.Duration) {
	if interval <= 0 {

	}
	tick := time.NewTicker(interval.Truncate(time.Second))
	for range tick.C {
		c.mu.Lock()
		for k, v := range c.entries {
			if time.Since(v.createdAt) > interval {
				delete(c.entries, k)
			}
		}
		c.mu.Unlock()
	}
}
