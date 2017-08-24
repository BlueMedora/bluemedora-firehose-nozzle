package ttlcache

import (
	"time"
	"sync"
)

type TTLCache struct {
	mutext sync.RWMutex
	TTL time.Duration
	resources map[string]*Resource
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

func (cache *TTLCache) Set(key string, resource *Resource) {
	cache.mutext.Lock()
	defer cache.mutext.Unlock()
	cache.resources[key] = resource
}

func (cache *TTLCache) Get(key string) (resource *Resource, found bool) {
	cache.mutext.RLock()
	defer cache.mutext.RUnlock()
	resource, found = cache.resources[key]
	return resource, found
}

func (cache *TTLCache) cleanup() {
	cache.mutext.Lock()
	defer cache.mutext.Unlock()

	for key, resource := range cache.resources {
		resource.cleanup()
		if resource.isEmpty() {
			delete(cache.resources, key)
		}
	}
}

func (cache *TTLCache) startCleanupTimer() {
	duration := cache.TTL
	if duration < time.Second {
		duration = time.Second
	}
	ticker := time.Tick(duration)
	go (func() {
		for {
			select {
				case <- ticker:
					cache.cleanup()
			}
		}
	}) ()
}

func createTTLCache() *TTLCache {
	cache := &TTLCache {
		resources: make(map[string]*Resource),
	}

	cache.startCleanupTimer()
	return cache
}
