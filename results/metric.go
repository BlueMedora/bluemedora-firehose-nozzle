package results

import (
	"sync"
	"time"
)

//Metric represents a single dropsonde metric
type Metric struct {
	sync.RWMutex
	data      float64
	timestamp int64
	expires   *time.Time
}

func (m *Metric) GetData() float64 {
	m.RLock()
	defer m.RUnlock()
	return m.data
}

func (m *Metric) GetTimestamp() int64 {
	m.RLock()
	defer m.RUnlock()
	return m.timestamp
}

func NewMetric(d float64, t int64, ttl time.Duration) *Metric {
	metric := &Metric{}
	metric.Update(d, t, ttl)
	return metric
}

func (m *Metric) Update(d float64, t int64, ttl time.Duration) {
	m.Lock()
	defer m.Unlock()
	expiration := time.Now().Add(ttl)
	m.expires = &expiration
	m.data = d
	m.timestamp = t
}

func (m *Metric) HasExpired() bool {
	m.RLock()
	defer m.RUnlock()
	if m.expires == nil {
		return true
	}

	return m.expires.Before(time.Now())
}
