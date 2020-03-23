package results

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/cloudfoundry/gosteno"
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
				ValueMetrics:   make(map[string][]*Metric),
				CounterMetrics: make(map[string][]*Metric),
			},
		},
	}

	for _, tc := range testCases {
		createdResource := NewResource(deployment, job, index, ip)

		if createdResource.deployment != tc.want.deployment || createdResource.job != tc.want.job || createdResource.index != tc.want.index || createdResource.ip != tc.want.ip {
			t.Errorf("Test Case %s returned %v expected %v", tc.testName, createdResource, tc.want)
		}
	}
}

func TestGetMetric(t *testing.T) {

	//Test not passing
	dummy := "dummy"
	resource := newTestResource()

	//Create new metric
	metrics := resource.getMetrics(resource.ValueMetrics, dummy)

	newMetric := resource.getMetrics(resource.ValueMetrics, dummy)
	if !reflect.DeepEqual(newMetric, metrics) {
		t.Errorf("Expecting %v, got %v", metrics, newMetric)
	}
}

func TestIsEmpty(t *testing.T) {
	resource := newTestResource()

	if !resource.IsEmpty() {
		t.Error("Resource was not empty")
	}

	resource.ValueMetrics["test"] = []*Metric{&Metric{}}

	if resource.IsEmpty() {
		t.Error("Resource was empty")
	}
}

func TestCleanup(t *testing.T) {
	expiration := time.Now().Add(time.Second)
	resource := newTestResource()

	resource.ValueMetrics["test"] = []*Metric{&Metric{expires: &expiration}, &Metric{expires: &expiration}}
	resource.CounterMetrics["test"] = []*Metric{&Metric{expires: &expiration}, &Metric{expires: &expiration}}

	resource.Cleanup()

	if resource.IsEmpty() {
		t.Error("Resource was empty after adding metrics")
	}

	//Sleep to allow expiration
	time.Sleep(2 * time.Second)

	resource.Cleanup()

	if !resource.IsEmpty() {
		t.Error("Resource was not empty after metrics expired")
	}
}

func TestRetainedDataAfterCleanup(t *testing.T) {
	expiration := time.Now().Add(time.Millisecond * 500)
	longerExpiration := time.Now().Add(time.Minute)
	resource := newTestResource()

	resource.ValueMetrics["test"] = []*Metric{&Metric{expires: &expiration}, &Metric{expires: &longerExpiration}}
	resource.CounterMetrics["test"] = []*Metric{&Metric{expires: &expiration}, &Metric{expires: &longerExpiration}}

	resource.Cleanup()

	if resource.IsEmpty() {
		t.Error("Resource was empty after adding metrics")
	}

	if len(resource.ValueMetrics["test"]) != 2 || len(resource.CounterMetrics["test"]) != 2 {
		t.Error("Metrics were not created for resource")
	}

	//Sleep to allow expiration
	time.Sleep(time.Second)

	resource.Cleanup()

	if len(resource.ValueMetrics["test"]) != 1 {
		t.Error("Incorrect number of value metrics remain after cleanup")
	}

	if len(resource.CounterMetrics["test"]) != 1 {
		t.Error("Incorrect number of counter metrics remain after cleanup")
	}
}

func TestAddMetric(t *testing.T) {
	origin, deployment, job, index, ip := "origin", "deployment", "job", "index", "ip"
	timestamp := time.Now().UnixNano()
	metricName, counterName := "metric", "counter"
	value, delta, total := float64(24), uint64(24), uint64(24)
	logger := createLogger()

	gaugeEnvelope := &loggregator_v2.Envelope{
		Timestamp:  timestamp,
		SourceId:   "sourceid",
		InstanceId: "instanceid",
		Tags: map[string]string{
			"deployment": deployment,
			"job":        job,
			"index":      index,
			"ip":         ip,
			"origin":     origin,
		},
		Message: &loggregator_v2.Envelope_Gauge{
			Gauge: &loggregator_v2.Gauge{
				Metrics: map[string]*loggregator_v2.GaugeValue{
					metricName: &loggregator_v2.GaugeValue{
						Unit:  "ms",
						Value: value,
					},
				},
			},
		},
	}

	counterEnvelope := &loggregator_v2.Envelope{
		Timestamp:  timestamp,
		SourceId:   "sourceid",
		InstanceId: "instanceid",
		Tags: map[string]string{
			"deployment": deployment,
			"job":        job,
			"index":      index,
			"ip":         ip,
			"origin":     origin,
		},
		Message: &loggregator_v2.Envelope_Counter{
			Counter: &loggregator_v2.Counter{
				Name:  counterName,
				Delta: delta,
				Total: total,
			},
		},
	}

	resource := newTestResource()
	ttl := 10 * time.Second

	//Test adding value metric
	resource.AddMetric(gaugeEnvelope, logger, ttl)

	if resource.IsEmpty() {
		t.Error("No metrics found in resource")
	}

	metrics := resource.getMetrics(resource.ValueMetrics, metricName)
	if metrics == nil || len(metrics) == 0 || metrics[0].data != value {
		t.Errorf("Metric %s not stored correctly", metricName)
	}

	delete(resource.ValueMetrics, metricName)

	//Test adding counter metric
	resource.AddMetric(counterEnvelope, logger, ttl)

	if resource.IsEmpty() {
		t.Error("No metrics found in resource")
	}

	metrics = resource.getMetrics(resource.CounterMetrics, counterName)
	if metrics == nil || len(metrics) == 0 || metrics[0].data != float64(total) {
		t.Errorf("Metric %s not stored correctly", counterName)
	}

	delete(resource.CounterMetrics, counterName)
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

	resource := newTestResource()

	resource.ValueMetrics["one"] = []*Metric{&Metric{data: 1, timestamp: int64(1257894000000000000)}}

	resource.CounterMetrics["one"] = []*Metric{&Metric{data: 1, timestamp: int64(1257894000000000000)}}

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

func newTestResource() *Resource {
	deployment, job, index, ip := "deployment", "job", "index", "ip"
	return NewResource(deployment, job, index, ip)
}
