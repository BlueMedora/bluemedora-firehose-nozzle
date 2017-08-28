package ttlcache

import (
	"sync"
	"time"
)

const cacheFlushSecond = 10

type TTLCache struct {
	mutext  sync.RWMutex
	TTL     time.Duration
	origins map[string]map[string]*Resource
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

func (cache *TTLCache) SetResource(originKey, key string, resource *Resource) {
	cache.mutext.Lock()
	defer cache.mutext.Unlock()

	var origin map[string]*Resource
	if value, ok := cache.origins[originKey]; ok {
		origin = value
	} else {
		origin = make(map[string]*Resource)
		cache.origins[originKey] = origin
	}

	origin[key] = resource
}

func (cache *TTLCache) GetResource(originKey, key string) (resource *Resource, found bool) {
	cache.mutext.RLock()
	defer cache.mutext.RUnlock()

	if origin, exists := cache.origins[originKey]; exists {
		resource, found = origin[key]
	}
	return resource, found
}

func (cache *TTLCache) GetOrigin(originKey string) (origin map[string]*Resource, found bool) {
	cache.mutext.RLock()
	defer cache.mutext.RUnlock()

	origin, found = cache.origins[originKey]
	return origin, found
}

func (cache *TTLCache) cleanup() {
	cache.mutext.Lock()
	defer cache.mutext.Unlock()

	for originKey, origin := range cache.origins {
		for key, resource := range origin {
			resource.cleanup()
			if resource.isEmpty() {
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
		origins: make(map[string]map[string]*Resource),
	}

	cache.startCleanupTimer()
	return cache
}
