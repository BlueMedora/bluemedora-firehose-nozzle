package ttlcache

import (
	"sync"
	
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/sonde-go/events"
)

//Resource represents cloud controller data
type Resource struct {
	mutext 		   sync.RWMutex
	Deployment     string
	Job            string
	Index          string
	IP             string
	valueMetrics   map[string]*Metric
	counterMetrics map[string]*Metric
}

func (r *Resource) AddMetric(envelope *events.Envelope, logger *gosteno.Logger) {
	var metric *Metric
	switch envelope.GetEventType() {
	case events.Envelope_ValueMetric:
		valueMetric := envelope.GetValueMetric()
		r.mutext.Lock()

		metric = r.getMetric(r.valueMetrics, valueMetric.GetName())
		metric.update(valueMetric.GetValue(), GetInstance().TTL)

		r.mutext.Unlock()
		logger.Debugf("Adding Value Event Name %s, Value %d", valueMetric.GetName(), valueMetric.GetValue())
	case events.Envelope_CounterEvent:
		counterEvent := envelope.GetCounterEvent()
		r.mutext.Lock()

		metric = r.getMetric(r.counterMetrics, counterEvent.GetName())
		metric.update(float64(counterEvent.GetTotal()), GetInstance().TTL)

		r.mutext.Unlock()
		logger.Debugf("Adding Counter Event Name %s, Value %d", counterEvent.GetName(), counterEvent.GetTotal())
	case events.Envelope_ContainerMetric:
		// ignored message type
	case events.Envelope_LogMessage:
		// ignored message type
	case events.Envelope_HttpStartStop:
		// ignored message type
	case events.Envelope_Error:
		// ignored message type
	default:
		logger.Warnf("Unknown event type %s", envelope.GetEventType())
	}
}

func (r *Resource) isEmpty() bool {
	r.mutext.RLock()
	defer r.mutext.RUnlock()
	count := len(r.valueMetrics)
	count += len(r.counterMetrics)
	return count == 0
}

func (r *Resource) cleanup() {
	r.mutext.Lock()
	defer r.mutext.Unlock()

	for key, metric := range r.valueMetrics {
		if metric.expired() {
			delete(r.valueMetrics, key)
		}
	}

	for key, metric := range r.counterMetrics {
		if metric.expired() {
			delete(r.counterMetrics, key)
		}
	}
}

func (r *Resource) getMetric(metricMap map[string]*Metric, metricName string) *Metric {
	var metric *Metric
	if metric, ok := metricMap[metricName]; !ok {
		metric = &Metric{}
		metricMap[metricName] = metric
	}

	return metric
}