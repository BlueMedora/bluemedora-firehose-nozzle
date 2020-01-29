package results

import (
	"encoding/json"
	"sync"

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
	valueMetrics   map[string][]*Metric
	counterMetrics map[string][]*Metric
}

//CreateResource Creates a new resource
func CreateResource(deployment, job, index, ip string) *Resource {
	return &Resource{
		deployment:     deployment,
		job:            job,
		index:          index,
		ip:             ip,
		valueMetrics:   make(map[string][]*Metric),
		counterMetrics: make(map[string][]*Metric),
	}
}

//AddMetric adds a metric to a resource
func (r *Resource) AddMetric(envelope *loggregator_v2.Envelope, logger *gosteno.Logger) {
	// var metrics []*Metric

	// timestamp := envelope.GetTimestamp()
    
    
	// switch envelope.Type() {
	// case loggregator_v2.Envelope_ValueMetric:
	// 	valueMetric := envelope.GetValueMetric()
	// 	r.mutext.Lock()

	// 	metrics = r.getMetrics(r.valueMetrics, valueMetric.GetName())
	// 	var metric = &Metric{}
	// 	metric.update(valueMetric.GetValue(), timestamp, GetInstance().TTL)
	// 	metrics = append(metrics, metric)
	// 	r.valueMetrics[valueMetric.GetName()] = metrics
	// 	r.mutext.Unlock()
	// 	logger.Debugf("Adding Value Event Name %s, Value %d", valueMetric.GetName(), valueMetric.GetValue())
	// case loggregator_v2.Envelope_CounterEvent:
	// 	counterEvent := envelope.GetCounterEvent()
	// 	r.mutext.Lock()

	// 	metrics = r.getMetrics(r.counterMetrics, counterEvent.GetName())
	// 	var metric = &Metric{}
	// 	metric.update(float64(counterEvent.GetTotal()), timestamp, GetInstance().TTL)
	// 	metrics = append(metrics, metric)
	// 	r.counterMetrics[counterEvent.GetName()] = metrics
	// 	r.mutext.Unlock()
	// 	logger.Debugf("Adding Counter Event Name %s, Value %d", counterEvent.GetName(), counterEvent.GetTotal())
	// case loggregator_v2.Envelope_ContainerMetric:
	// 	// ignored message type
	// case loggregator_v2.Envelope_LogMessage:
	// 	// ignored message type
	// case loggregator_v2.Envelope_HttpStartStop:
	// 	// ignored message type
	// case loggregator_v2.Envelope_Error:
	// 	// ignored message type
	// default:
	// 	logger.Warnf("Unknown event type %s", envelope.GetEventType())
	// }
}

func parseGauge(e *loggregator_v2.Envelope_Gauge) *loggregator_v2.Envelope_Gauge {
	if e != nil {
		return e
	}
    return nil

	// valueMetric := envelope.GetValueMetric()
	// 	r.mutext.Lock()

	// 	metrics = r.getMetrics(r.valueMetrics, valueMetric.GetName())
	// 	var metric = &Metric{}
	// 	metric.update(valueMetric.GetValue(), timestamp, GetInstance().TTL)
	// 	metrics = append(metrics, metric)
	// 	r.valueMetrics[valueMetric.GetName()] = metrics
	// 	r.mutext.Unlock()
	// 	logger.Debugf("Adding Value Event Name %s, Value %d", valueMetric.GetName(), valueMetric.GetValue())
}


func (r *Resource) IsEmpty() bool {
	r.mutext.RLock()
	defer r.mutext.RUnlock()
	count := 0
	for _, metrics := range r.valueMetrics {
		count += len(metrics)
	}
	// count := len(r.valueMetrics)
	for _, metrics := range r.counterMetrics {
		count += len(metrics)
	}
	return count == 0
}

func (r *Resource) Cleanup() {
	r.mutext.Lock()
	defer r.mutext.Unlock()

	for key, metrics := range r.valueMetrics {
		r.valueMetrics[key] = nonExpiredMetric(metrics)
	}

	for key, metrics := range r.counterMetrics {
		r.counterMetrics[key] = nonExpiredMetric(metrics)
	}
}

func nonExpiredMetric(metrics []*Metric) []*Metric {
	var metricsToKeep []*Metric
	for _, metric := range metrics {
		if !metric.expired() {
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
	valueMetrics, counterMetrics := convertMap(r.valueMetrics), convertMap(r.counterMetrics)

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
		ValueMetrics:   valueMetrics,
		CounterMetrics: counterMetrics,
	})
}

func convertMap(inputMap map[string][]*Metric) map[string]metricsJSON {
	outputMap := make(map[string]metricsJSON)
	for key, metrics := range inputMap {
		var emptyList []metricJSON
		var jsonMetrics = metricsJSON{Metrics: emptyList}
		for _, metric := range metrics {
			jsonMetrics.Metrics = append(jsonMetrics.Metrics, metricJSON{Value: metric.getData(), Timestamp: metric.getTimestamp()})
		}
		outputMap[key] = jsonMetrics
	}
	return outputMap
}
