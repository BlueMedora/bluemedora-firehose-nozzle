package ttlcache

import (
	"os"
	"testing"
	"time"

	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/sonde-go/events"
)

func TestCreateResource(t *testing.T) {
	deployment, job, index, ip := "deployment", "job", "index", "ip"

	testCases := []struct {
		testName string
		want     *Resource
	}{
		{
			testName: "Normal Creation",
			want: &Resource{
				deployment:     deployment,
				job:            job,
				index:          index,
				ip:             ip,
				valueMetrics:   make(map[string]*Metric),
				counterMetrics: make(map[string]*Metric),
			},
		},
	}

	for _, tc := range testCases {
		createdResource := CreateResource(deployment, job, index, ip)

		if createdResource.deployment != tc.want.deployment || createdResource.job != tc.want.job || createdResource.index != tc.want.index || createdResource.ip != tc.want.ip {
			t.Errorf("Test Case %s returned %v expected %v", tc.testName, createdResource, tc.want)
		}
	}
}

func TestGetMetric(t *testing.T) {

	//Test not passing
	dummy := "dummy"
	resource := createTestResource()

	//Create new metric
	metric := resource.getMetric(resource.valueMetrics, dummy)

	if metric == nil {
		t.Error("No new metric created")
	}

	newMetric := resource.getMetric(resource.valueMetrics, dummy)
	if newMetric != metric {
		t.Errorf("Expecting %v, got %v", metric, newMetric)
	}
}

func TestIsEmpty(t *testing.T) {
	resource := createTestResource()

	if !resource.isEmpty() {
		t.Error("Resource was not empty")
	}

	resource.valueMetrics["test"] = &Metric{}

	if resource.isEmpty() {
		t.Error("Resource was empty")
	}
}

func TestCleanup(t *testing.T) {
	expiration := time.Now().Add(time.Second)
	resource := createTestResource()

	resource.valueMetrics["test"] = &Metric{expires: &expiration}
	resource.counterMetrics["test"] = &Metric{expires: &expiration}

	resource.cleanup()

	if resource.isEmpty() {
		t.Error("Resource was empty after adding metrics")
	}

	//Sleep to allow expiration
	time.Sleep(2 * time.Second)

	resource.cleanup()
	
	if !resource.isEmpty() {
		t.Error("Resource was not empty after metrics expired")
	}
}

func TestAddMetric(t *testing.T) {
	origin, deployment, job, index, ip := "origin", "deployment", "job", "index", "ip"
	metricName, counterName := "metric", "counter"
	value, delta, total := float64(24), uint64(24), uint64(24)
	valueType, counterType := events.Envelope_ValueMetric, events.Envelope_CounterEvent
	logger := createLogger()
	GetInstance().TTL = time.Second

	metricEnvelope := &events.Envelope{
		Origin:     &origin,
		Deployment: &deployment,
		Job:        &job,
		Index:      &index,
		Ip:         &ip,
		EventType:  &valueType,
		ValueMetric: &events.ValueMetric{
			Name:  &metricName,
			Value: &value,
		},
	}

	counterEnvelope := &events.Envelope{
		Origin:     &origin,
		Deployment: &deployment,
		Job:        &job,
		Index:      &index,
		Ip:         &ip,
		EventType:  &counterType,
		CounterEvent: &events.CounterEvent{
			Name:  &counterName,
			Delta: &delta,
			Total: &total,
		},
	}

	resource := createTestResource()

	//Test adding value metric
	resource.AddMetric(metricEnvelope, logger)

	if resource.isEmpty() {
		t.Error("No metrics found in resource")
	}

	metric := resource.getMetric(resource.valueMetrics, metricName)
	if metric == nil || metric.data != value {
		t.Errorf("Metric %s not stored correctly", metricName)
	}

	delete(resource.valueMetrics, metricName)

	//Test adding counter metric
	resource.AddMetric(counterEnvelope, logger)

	if resource.isEmpty() {
		t.Error("No metrics found in resource")
	}

	metric = resource.getMetric(resource.counterMetrics, counterName)
	if metric == nil || metric.data != float64(total) {
		t.Errorf("Metric %s not stored correctly", counterName)
	}

	delete(resource.valueMetrics, counterName)
}

func TestConvertMap(t *testing.T) {
	testCases := []struct {
		testName string
		input	 map[string]*Metric
		want     map[string]float64
	}{
		{
			testName: "Blank Input",
			input: make(map[string]*Metric),
			want: make(map[string]float64),
		},
		{
			testName: "Normal Input",
			input: map[string]*Metric {
				"one": &Metric{data: 1},
				"two": &Metric{data: 2},
				"three": &Metric{data: 3},
			},
			want: map[string]float64 {
				"one": 1,
				"two": 2,
				"three": 3,
			},
		},
	}

	for _, tc := range testCases {
		output := convertMap(tc.input)

		equal := true
		for key, value := range tc.want {
			if outputVal, ok := output[key]; !ok {
				equal = false
			} else if outputVal != value {
				equal = false
			}
		}

		if !equal {
			t.Errorf("Got %v expected %v", output, tc.want)
		}
	}
}

func createLogger() *gosteno.Logger {
	//Ceate logger
	config := &gosteno.Config{
		Sinks:     make([]gosteno.Sink, 1),
		Level:     gosteno.LOG_INFO,
		Codec:     gosteno.NewJsonCodec(),
		EnableLOC: true,
	}

	config.Sinks[0] = gosteno.NewIOSink(os.Stdout)

	gosteno.Init(config)
	return gosteno.NewLogger("logger")
}

func createTestResource() *Resource {
	deployment, job, index, ip := "deployment", "job", "index", "ip"
	return CreateResource(deployment, job, index, ip)
}
