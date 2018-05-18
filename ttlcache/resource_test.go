package ttlcache

import (
	"encoding/json"
	"os"
	"reflect"
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
				valueMetrics:   make(map[string][]*Metric),
				counterMetrics: make(map[string][]*Metric),
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
	metrics := resource.getMetrics(resource.valueMetrics, dummy)

	newMetric := resource.getMetrics(resource.valueMetrics, dummy)
	if !reflect.DeepEqual(newMetric, metrics) {
		t.Errorf("Expecting %v, got %v", metrics, newMetric)
	}
}

func TestIsEmpty(t *testing.T) {
	resource := createTestResource()

	if !resource.isEmpty() {
		t.Error("Resource was not empty")
	}

	resource.valueMetrics["test"] = []*Metric{&Metric{}}

	if resource.isEmpty() {
		t.Error("Resource was empty")
	}
}

func TestCleanup(t *testing.T) {
	expiration := time.Now().Add(time.Second)
	resource := createTestResource()

	resource.valueMetrics["test"] = []*Metric{&Metric{expires: &expiration}, &Metric{expires: &expiration}}
	resource.counterMetrics["test"] = []*Metric{&Metric{expires: &expiration}, &Metric{expires: &expiration}}

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

func TestRetainedDataAfterCleanup(t *testing.T) {
	expiration := time.Now().Add(time.Millisecond * 500)
	longerExpiration := time.Now().Add(time.Minute)
	resource := createTestResource()

	resource.valueMetrics["test"] = []*Metric{&Metric{expires: &expiration}, &Metric{expires: &longerExpiration}}
	resource.counterMetrics["test"] = []*Metric{&Metric{expires: &expiration}, &Metric{expires: &longerExpiration}}

	resource.cleanup()

	if resource.isEmpty() {
		t.Error("Resource was empty after adding metrics")
	}

	if len(resource.valueMetrics["test"]) != 2 || len(resource.counterMetrics["test"]) != 2 {
		t.Error("Metrics were not created for resource")
	}

	//Sleep to allow expiration
	time.Sleep(time.Second)

	resource.cleanup()

	if len(resource.valueMetrics["test"]) != 1 {
		t.Error("Incorrect number of value metrics remain after cleanup")
	}

	if len(resource.counterMetrics["test"]) != 1 {
		t.Error("Incorrect number of counter metrics remain after cleanup")
	}
}

func TestAddMetric(t *testing.T) {
	origin, deployment, job, index, ip := "origin", "deployment", "job", "index", "ip"
	timestamp := time.Now().UnixNano()
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
		Timestamp:  &timestamp,
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
		Timestamp:  &timestamp,
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

	metrics := resource.getMetrics(resource.valueMetrics, metricName)
	if metrics == nil || len(metrics) == 0 || metrics[0].data != value {
		t.Errorf("Metric %s not stored correctly", metricName)
	}

	delete(resource.valueMetrics, metricName)

	//Test adding counter metric
	resource.AddMetric(counterEnvelope, logger)

	if resource.isEmpty() {
		t.Error("No metrics found in resource")
	}

	metrics = resource.getMetrics(resource.counterMetrics, counterName)
	if metrics == nil || len(metrics) == 0 || metrics[0].data != float64(total) {
		t.Errorf("Metric %s not stored correctly", counterName)
	}

	delete(resource.valueMetrics, counterName)
}

func TestConvertMap(t *testing.T) {
	testCases := []struct {
		testName string
		input    map[string][]*Metric
		want     map[string]metricsJSON
	}{
		{
			testName: "Blank Input",
			input:    make(map[string][]*Metric),
			want:     make(map[string]metricsJSON),
		},
		{
			testName: "Normal Input",
			input: map[string][]*Metric{
				"one":   []*Metric{&Metric{data: 1, timestamp: int64(1257894000000000000)}},
				"two":   []*Metric{&Metric{data: 2, timestamp: int64(1257894000000000000)}},
				"three": []*Metric{&Metric{data: 3, timestamp: int64(1257894000000000000)}},
			},
			want: map[string]metricsJSON{
				"one":   metricsJSON{[]metricJSON{metricJSON{Value: 1, Timestamp: int64(1257894000000000000)}}},
				"two":   metricsJSON{[]metricJSON{metricJSON{Value: 2, Timestamp: int64(1257894000000000000)}}},
				"three": metricsJSON{[]metricJSON{metricJSON{Value: 3, Timestamp: int64(1257894000000000000)}}},
			},
		},
	}

	for _, tc := range testCases {
		output := convertMap(tc.input)

		equal := true
		for key, value := range tc.want {
			if outputVal, ok := output[key]; !ok {
				equal = false
			} else if !reflect.DeepEqual(outputVal, value) {
				equal = false
			}
		}

		if !equal {
			t.Errorf("Got %v expected %v", output, tc.want)
		}
	}
}

func TestMarshalJSON(t *testing.T) {
	want := `{"Deployment":"deployment","Job":"job","Index":"index","IP":"ip","ValueMetrics":{"one":{"metrics":[{"value":1,"timestamp":1257894000000000000}]}},"CounterMetrics":{"one":{"metrics":[{"value":1,"timestamp":1257894000000000000}]}}}`

	resource := createTestResource()

	resource.valueMetrics["one"] = []*Metric{&Metric{data: 1, timestamp: int64(1257894000000000000)}}

	resource.counterMetrics["one"] = []*Metric{&Metric{data: 1, timestamp: int64(1257894000000000000)}}

	messageBytes, err := resource.MarshalJSON()
	if err != nil {
		t.Errorf("Error marshalling json %s", err.Error())
	}

	jsonString := string(messageBytes)

	if jsonString != want {
		t.Errorf("Failed testing Marshall directly, expecting %s\n got %s", want, jsonString)
	}

	//Testing using json package
	messageBytes, err = json.Marshal(resource)
	if err != nil {
		t.Errorf("Error marshalling json %s", err.Error())
	}

	jsonString = string(messageBytes)
	if jsonString != want {
		t.Errorf("Failed testing with json package, expecting %s\n got %s", want, jsonString)
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
