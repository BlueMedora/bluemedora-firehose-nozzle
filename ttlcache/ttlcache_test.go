package ttlcache

import (
	"testing"
	"time"

	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/results"
)

func TestGetInstance(t *testing.T) {
	instancePointer := GetInstance()

	secondPointer := GetInstance()

	//Test to ensure pointers point at the same object
	if instancePointer != secondPointer {
		t.Errorf("Expecting %v got %v", instancePointer, secondPointer)
	}
}

func TestCreateTTLCache(t *testing.T) {
	testCases := []struct {
		testName string
		want     *TTLCache
	}{
		{
			testName: "Normal Creation",
			want: &TTLCache{
				origins: make(map[string]map[string]*results.Resource),
			},
		},
	}

	for _, tc := range testCases {
		createdCache := createTTLCache()

		if createdCache.TTL != tc.want.TTL || createdCache.origins == nil {
			t.Errorf("Test Case %s returned %v expected %v", tc.testName, createdCache, tc.want)
		}
	}
}

func TestSetResource(t *testing.T) {
	resource := &results.Resource{}
	origin, key := "origin", "key"

	cache := &TTLCache{origins: make(map[string]map[string]*results.Resource)}

	cache.SetResource(origin, key, resource)

	if originMap, ok := cache.origins[origin]; ok {
		getResource := originMap[key]

		if getResource == nil || getResource != resource {
			t.Errorf("Expecting %v got %v", resource, getResource)
		}
	} else {
		t.Errorf("Origin map did not exists")
	}
}

func TestGetResource(t *testing.T) {
	resource := &results.Resource{}
	origin, key := "origin", "key"

	cache := &TTLCache{origins: make(map[string]map[string]*results.Resource)}

	//Testing with empty cache
	if _, found := cache.GetResource(origin, key); found {
		t.Error("Found resource that didn't exists")
	}

	//Add resource to cache
	cache.SetResource(origin, key, resource)

	if getResource, found := cache.GetResource(origin, key); found {
		if getResource != resource {
			t.Errorf("Expecting: %v got: %v", resource, getResource)
		}
	} else {
		t.Error("No resource found in non-empty cache")
	}
}

func TestGetOrigin(t *testing.T) {
	resource := &results.Resource{}
	origin := "origin"
	originMap := map[string]*results.Resource{
		"key": resource,
	}
	cache := &TTLCache{origins: make(map[string]map[string]*results.Resource)}

	//Test empty cache
	if _, found := cache.GetOrigin(origin); found {
		t.Error("Found origin in empty cache")
	}

	//Test Non empty cache
	cache.origins[origin] = originMap

	if getOriginMap, found := cache.GetOrigin(origin); found {
		for key, value := range originMap {
			getValue := getOriginMap[key]
			if getValue != value {
				t.Errorf("Expecting: %v, got: %v", originMap, getOriginMap)
			}
		}
	} else {
		t.Error("No origin map found")
	}
}

func TestCacheCleanup(t *testing.T) {
	expiration := time.Second * 1

	cache := &TTLCache{
		origins: make(map[string]map[string]*results.Resource),
	}

	resource := newTestResource()
	m1 := &results.Metric{}
	m1.Update(2,2,expiration)
	m2 := &results.Metric{}
	m2.Update(4,4,expiration)
	resource.ValueMetrics["test"] = append(resource.ValueMetrics["test"],  m1)
	resource.ValueMetrics["test"] = append(resource.ValueMetrics["test"],  m2)


	cache.SetResource("origin", "key", resource)

	cache.cleanup()

	if len(cache.origins) != 1 {
		t.Error("Cache cleaned up before expiration")
	}

	time.Sleep(2 * time.Second)
	cache.cleanup()

	if len(cache.origins) != 0 {
		t.Error("Failed to fully clean out cache after expiration")
	}
}

func newTestResource() *results.Resource {
	deployment, job, index, ip := "deployment", "job", "index", "ip"
	return results.NewResource(deployment, job, index, ip)
}