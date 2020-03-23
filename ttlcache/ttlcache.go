package ttlcache

import (
	"fmt"
	"sync"
	"time"

	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/results"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/cloudfoundry/gosteno"
)

const (
	cacheFlushInterval = 10 * time.Second
)

type TTLCache struct {
	sync.RWMutex
	TTL     time.Duration
	logger  *gosteno.Logger
	origins map[string]map[string]*results.Resource
}

var instance *TTLCache
var once sync.Once

//GetInstance retrieves the singleton cache
func GetInstance() *TTLCache {
	return instance
}

func CreateInstance(logger *gosteno.Logger) {
	once.Do(func() {
		if logger == nil {
			panic("Cache initialized without logger")
		}
		instance = createTTLCache(logger)
	})
}

// todo channel for storing messages and having time to update this without locking reading
func (c *TTLCache) UpdateResource(e *loggregator_v2.Envelope) {
	c.Lock()
	defer c.Unlock()

	k := createEnvelopeKey(e)
	var r *results.Resource
	if value, ok := c.getResource(e.Tags["origin"], k); ok {
		r = value
	} else {
		r = results.NewResource(e.Tags["deployment"], e.Tags["job"], e.Tags["index"], e.Tags["ip"])
		c.setResource(e.Tags["origin"], k, r)
	}

	r.AddMetric(e, c.logger, c.TTL)
}

func createEnvelopeKey(e *loggregator_v2.Envelope) string {
	return fmt.Sprintf("%s | %s | %s | %s", e.Tags["deployment"], e.Tags["job"], e.Tags["index"], e.Tags["ip"])
}

// private utility func, public methods using it are expected to have mutex lock
func (c *TTLCache) setResource(originKey, key string, resource *results.Resource) {
	var origin map[string]*results.Resource
	if value, ok := c.origins[originKey]; ok {
		origin = value
	} else {
		origin = make(map[string]*results.Resource)
		c.origins[originKey] = origin
	}

	origin[key] = resource
}

func (c *TTLCache) GetResource(originKey, key string) (resource *results.Resource, found bool) {
	c.RLock()
	defer c.RUnlock()
	return c.getResource(originKey, key)
}

// private utility func, public methods using it are expected to have mutex RLock minimum
func (c *TTLCache) getResource(originKey, key string) (resource *results.Resource, found bool) {
	if origin, exists := c.origins[originKey]; exists {
		resource, found = origin[key]
	}
	return resource, found
}

func (c *TTLCache) GetOrigin(originKey string) (origin map[string]*results.Resource, found bool) {
	c.logger.Info("Get Origin")
	c.RLock()
	defer c.RUnlock()

	origin, found = c.origins[originKey]
	c.logger.Info("Returning from origin")
	return origin, found
}

func (c *TTLCache) cleanup() {
	c.Lock()
	defer c.Unlock()

	for originKey, origin := range c.origins {
		for key, resource := range origin {
			resource.Cleanup()
			if resource.IsEmpty() {
				delete(origin, key)
			}
		}

		if len(origin) == 0 {
			delete(c.origins, originKey)
		}
	}
}

func (c *TTLCache) startCleanupTimer() {
	duration := time.Duration(cacheFlushInterval)
	if duration < time.Second {
		duration = time.Second
	}
	ticker := time.Tick(duration)
	go (func() {
		for {
			select {
			case <-ticker:
				c.cleanup()
			}
		}
	})()
}

func createTTLCache(logger *gosteno.Logger) *TTLCache {
	c := &TTLCache{
		origins: make(map[string]map[string]*results.Resource),
		logger:  logger,
	}
	c.logger.Info("Built Cache")

	c.startCleanupTimer()
	return c
}
