package ttlcache

import (
	"testing"
	"time"
)

func TestExpired(t *testing.T) {
	metric := &Metric{data: 24}

	expiration := time.Now().Add(time.Second)
	metric.expires = &expiration
	if metric.expired() {
		t.Error("Expected metric to not be expired")
	}

	expiration = time.Now().Add(0 - time.Second)
	metric.expires = &expiration
	if !metric.expired() {
		t.Error("Expected metric to be expired")
	}
}

func TestGetData(t *testing.T) {
	data := float64(24)
	metric := &Metric{data: data}

	metricData := metric.getData()
	if metricData != data {
		t.Errorf("Expected %f got %f", data, metricData)
	}
}

func TestUpdate(t *testing.T) {
	metric := &Metric{}
	metric.update(35, time.Second)
	if metric.expired() {
		t.Error("Expected item to not be expired after update")
	}
}
