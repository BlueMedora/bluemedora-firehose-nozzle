package ttlcache

import (
	"sync"
	"time"
	"fmt"

	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/results"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/cloudfoundry/gosteno"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/logger"
)

const (
	cacheFlushSecond = 10
	logDirectory = "./logs"
	logFile = "bm_cache.log"
	logName = "bm_cache"
)

type TTLCache struct {
	mutext  sync.RWMutex
	TTL     time.Duration
	logger *gosteno.Logger
	origins map[string]map[string]*results.Resource
}

var instance *TTLCache
var once sync.Once

//GetInstance retrieves the singleton cache
func GetInstance() *TTLCache {
	once.Do(func() {
		instance = createTTLCache()
	})

	return instance
}

// todo channel for storing messages and having time to update this without locking reading
func (cache *TTLCache) UpdateResource(e *loggregator_v2.Envelope){
	cache.mutext.Lock()
	defer cache.mutext.Unlock()

	k := createEnvelopeKey(e)
	var r *results.Resource
	if value, ok := cache.getResource(e.Tags["origin"], k); ok {
		r = value
	} else {
		r = results.NewResource(e.Tags["deployment"], e.Tags["job"], e.Tags["index"], e.Tags["ip"])
		cache.setResource(e.Tags["origin"], k, r)
	}

	r.AddMetric(e, cache.logger, cache.TTL)
}

func createEnvelopeKey(e *loggregator_v2.Envelope) string {
	return fmt.Sprintf("%s | %s | %s | %s", e.Tags["deployment"], e.Tags["job"], e.Tags["index"], e.Tags["ip"])
}

// private utility func, public methods using it are expected to have mutext lock
func (cache *TTLCache) setResource(originKey, key string, resource *results.Resource) {
	var origin map[string]*results.Resource
	if value, ok := cache.origins[originKey]; ok {
		origin = value
	} else {
		origin = make(map[string]*results.Resource)
		cache.origins[originKey] = origin
	}

	origin[key] = resource
}

func (cache *TTLCache) GetResource(originKey, key string) (resource *results.Resource, found bool) {
	cache.mutext.RLock()
	defer cache.mutext.RUnlock()
	return cache.getResource(originKey, key)
}

// private utility func, public methods using it are expected to have mutext RLock minimum
func (cache *TTLCache) getResource(originKey, key string) (resource *results.Resource, found bool) {
	if origin, exists := cache.origins[originKey]; exists {
		resource, found = origin[key]
	}
	return resource, found
}

func (cache *TTLCache) GetOrigin(originKey string) (origin map[string]*results.Resource, found bool) {
	cache.logger.Info("Get Origin")
	cache.mutext.RLock()
	defer cache.mutext.RUnlock()

	origin, found = cache.origins[originKey]
	cache.logger.Info("Returning from origin")
	return origin, found
}

func (cache *TTLCache) cleanup() {
	cache.mutext.Lock()
	defer cache.mutext.Unlock()

	for originKey, origin := range cache.origins {
		for key, resource := range origin {
			resource.Cleanup()
			if resource.IsEmpty() {
				delete(origin, key)
			}
		}

		if len(origin) == 0 {
			delete(cache.origins, originKey)
		}
	}
}

func (cache *TTLCache) startCleanupTimer() {
	duration := time.Duration(cacheFlushSecond) * time.Second
	if duration < time.Second {
		duration = time.Second
	}
	ticker := time.Tick(duration)
	go (func() {
		for {
			select {
			case <-ticker:
				cache.cleanup()
			}
		}
	})()
}

func createTTLCache() *TTLCache {
	cache := &TTLCache{
		origins: make(map[string]map[string]*results.Resource),
		logger: logger.New(logDirectory, logFile, logName, "info"),
	}
	cache.logger.Info("Built Cache")

	cache.startCleanupTimer()
	return cache
}
