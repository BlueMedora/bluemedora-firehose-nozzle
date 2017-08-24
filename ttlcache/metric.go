package ttlcache

import (
	"sync"
	"time"
)

//Metric represents a single dropsonde metric
type Metric struct {
	sync.RWMutex
	data float64
	expires *time.Time
}

func (m *Metric) getData() float64 {
	m.RLock()
	defer m.RUnlock()
	return m.data
}

func (m *Metric) update(newData float64, duration time.Duration) {
	m.Lock()
	defer m.Unlock()
	expiration := time.Now().Add(duration)
	m.expires = &expiration
	m.data = newData
}

func (m *Metric) expired() bool {
	m.RLock()
	defer m.RUnlock()
	if m.expires == nil {
		return true
	} 
	
	return m.expires.Before(time.Now())
}

