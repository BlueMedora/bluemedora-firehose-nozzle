package results

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/cloudfoundry/gosteno"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

//Resource represents cloud controller data
type Resource struct {
	mutext         sync.RWMutex
	deployment     string
	job            string
	index          string
	ip             string
	ValueMetrics   map[string][]*Metric
	CounterMetrics map[string][]*Metric
}

//CreateResource Creates a new resource
func NewResource(deployment, job, index, ip string) *Resource {
	return &Resource{
		deployment:     deployment,
		job:            job,
		index:          index,
		ip:             ip,
		ValueMetrics:   make(map[string][]*Metric),
		CounterMetrics: make(map[string][]*Metric),
	}
}

func (r *Resource) AddMetric(e *loggregator_v2.Envelope, l *gosteno.Logger, ttl time.Duration) {
	t := e.GetTimestamp()
    
    g := e.GetGauge()
    if g != nil {
    	r.addGaugeMetrics(g, l, t, ttl)
    }

    c := e.GetCounter()
    if c != nil {
    	r.addCounterMetric(c, l, t, ttl)
    }
}

func (r *Resource) addCounterMetric ( c *loggregator_v2.Counter, l *gosteno.Logger, timestamp int64, ttl time.Duration) {
	r.mutext.Lock()
	defer r.mutext.Unlock()
	r.CounterMetrics[c.GetName()] = append(
		r.getMetrics(r.CounterMetrics, c.GetName()),
		NewMetric(float64(c.GetTotal()), timestamp, ttl),
	)
	l.Debugf("Adding Value Event Name %s, Value %d", c.GetName(), c.GetTotal())    
}


func (r *Resource) addGaugeMetrics ( g *loggregator_v2.Gauge, l *gosteno.Logger, timestamp int64, ttl time.Duration) {
	r.mutext.Lock()
	defer r.mutext.Unlock()
	for k, v := range g.Metrics {
        r.ValueMetrics[k] = append(r.ValueMetrics[k], NewMetric(v.GetValue(), timestamp, ttl))
        l.Debugf("Adding Value Event Name %s, Value %f", k, v.GetValue())    
    }
}


func (r *Resource) IsEmpty() bool {
	r.mutext.RLock()
	defer r.mutext.RUnlock()
	count := 0
	for _, metrics := range r.ValueMetrics {
		count += len(metrics)
	}
	// count := len(r.valueMetrics)
	for _, metrics := range r.CounterMetrics {
		count += len(metrics)
	}
	return count == 0
}

func (r *Resource) Cleanup() {
	r.mutext.Lock()
	defer r.mutext.Unlock()

	for key, metrics := range r.ValueMetrics {
		r.ValueMetrics[key] = nonExpiredMetric(metrics)
	}

	for key, metrics := range r.CounterMetrics {
		r.CounterMetrics[key] = nonExpiredMetric(metrics)
	}
}

func nonExpiredMetric(metrics []*Metric) []*Metric {
	var metricsToKeep []*Metric
	for _, metric := range metrics {
		if !metric.HasExpired() {
			metricsToKeep = append(metricsToKeep, metric)
		}
	}
	return metricsToKeep
}

func (r *Resource) getMetrics(metricMap map[string][]*Metric, metricName string) []*Metric {
	var metrics []*Metric
	if value, ok := metricMap[metricName]; ok {
		return value
	}

	metricMap[metricName] = metrics
	return metrics
}

//metricJSON is a private struct for structure metrics in JSON
type metricJSON struct {
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
}

//metricsJSON is a struct to make a slice out of the metrics
type metricsJSON struct {
	Metrics []metricJSON `json:"metrics"`
}

func (r *Resource) MarshalJSON() ([]byte, error) {
	ValueMetrics, CounterMetrics := convertMap(r.ValueMetrics), convertMap(r.CounterMetrics)

	return json.Marshal(&struct {
		Deployment     string
		Job            string
		Index          string
		IP             string
		ValueMetrics   map[string]metricsJSON
		CounterMetrics map[string]metricsJSON
	}{
		Deployment:     r.deployment,
		Job:            r.job,
		Index:          r.index,
		IP:             r.ip,
		ValueMetrics:   ValueMetrics,
		CounterMetrics: CounterMetrics,
	})
}

func convertMap(inputMap map[string][]*Metric) map[string]metricsJSON {
	outputMap := make(map[string]metricsJSON)
	for key, metrics := range inputMap {
		var emptyList []metricJSON
		var jsonMetrics = metricsJSON{Metrics: emptyList}
		for _, metric := range metrics {
			jsonMetrics.Metrics = append(jsonMetrics.Metrics, metricJSON{Value: metric.GetData(), Timestamp: metric.GetTimestamp()})
		}
		outputMap[key] = jsonMetrics
	}
	return outputMap
}
