package results

import (
	"testing"
	"time"
)

func TestExpired(t *testing.T) {
	metric := &Metric{data: 24}

	expiration := time.Now().Add(time.Second)
	metric.expires = &expiration
	if metric.HasExpired() {
		t.Error("Expected metric to not be expired")
	}

	expiration = time.Now().Add(0 - time.Second)
	metric.expires = &expiration
	if !metric.HasExpired() {
		t.Error("Expected metric to be expired")
	}
}

func TestGetData(t *testing.T) {
	data := float64(24)
	metric := &Metric{data: data}

	metricData := metric.GetData()
	if metricData != data {
		t.Errorf("Expected %f got %f", data, metricData)
	}
}
func TestGetTimestamp(t *testing.T) {
	timestamp := time.Now().UnixNano()
	metric := &Metric{timestamp: timestamp}

	metricTimestamp := metric.GetTimestamp()
	if metricTimestamp != timestamp {
		t.Errorf("Expected %d got %d", timestamp, metricTimestamp)
	}
}

func TestUpdate(t *testing.T) {
	metric := &Metric{}
	timestamp := time.Now().UnixNano()
	data := float64(35)

	metric.Update(data, timestamp, time.Second)
	if metric.HasExpired() {
		t.Error("Expected item to not be expired after update")
	} else if metricData := metric.data; metricData != data {
		t.Errorf("Expected data %f got %f", data, metricData)
	} else if metricTime := metric.timestamp; metricTime != timestamp {
		t.Errorf("Expected timestamp %d got %d", timestamp, metricTime)
	}
}
