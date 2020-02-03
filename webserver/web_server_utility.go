// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package webserver

import (
	"fmt"
	"path/filepath"

    "github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/results"
	
	"github.com/cloudfoundry/gosteno"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

const (
	metronAgentOrigin       = "MetronAgent"
	syslogDrainBinderOrigin = "syslog_drain_binder"
	tpsWatcherOrigin        = "tps_watcher"
	tpsListenerOrigin       = "tps_listener"
	stagerOrigin            = "stager"
	sshProxyOrigin          = "ssh-proxy"
	senderOrigin            = "sender"
	routeEmitterOrigin      = "route_emitter"
	repOrigin               = "rep"
	receptorOrigin          = "receptor"
	nsyncListenerOrigin     = "nsync_listener"
	nsyncBulkerOrigin       = "nsync_bulker"
	gardenLinuxOrigin       = "garden-linux"
	fileServerOrigin        = "file_server"
	fetcherOrigin           = "fetcher"
	convergerOrigin         = "converger"
	ccUploaderOrigin        = "cc_uploader"
	bbsOrigin               = "bbs"
	auctioneerOrigin        = "auctioneer"
	etcdOrigin              = "etcd"
	dopplerServerOrigin     = "DopplerServer"
	cloudControllerOrigin   = "cc"
	trafficControllerOrigin = "LoggregatorTrafficController"
	goRouterOrigin          = "gorouter"
	locketOrigin            = "locket"
)

//Resource represents cloud controller data
type Resource struct {
	Deployment     string
	Job            string
	Index          string
	IP             string
	ValueMetrics   map[string]float64
	CounterMetrics map[string]float64
}

func createEnvelopeKey(e *loggregator_v2.Envelope) string {
	return fmt.Sprintf("%s | %s | %s | %s", e.Tags["deployment"], e.Tags["job"], e.Tags["index"], e.Tags["ip"])
}

func addMetric(e *loggregator_v2.Envelope, valueMetricMap map[string]float64, counterMetricMap map[string]float64, logger *gosteno.Logger) {
	// switch e.GetEventType() {
	// case events.Envelope_ValueMetric:
	// 	valueMetric := e.GetValueMetric()

	// 	valueMetricMap[valueMetric.GetName()] = valueMetric.GetValue()
	// 	logger.Debugf("Adding Value Event Name %s, Value %d", valueMetric.GetName(), valueMetric.GetValue())
	// case events.Envelope_CounterEvent:
	// 	counterEvent := envelope.GetCounterEvent()

	// 	counterMetricMap[counterEvent.GetName()] = float64(counterEvent.GetTotal())
	// 	logger.Debugf("Adding Counter Event Name %s, Value %d", counterEvent.GetName(), counterEvent.GetTotal())
	// case events.Envelope_ContainerMetric:
	// 	// ignored message type
	// case events.Envelope_LogMessage:
	// 	// ignored message type
	// case events.Envelope_HttpStartStop:
	// 	// ignored message type
	// case events.Envelope_Error:
	// 	// ignored message type
	// default:
	// 	logger.Warnf("Unknown event type %s", e.GetEventType())
	// }
}

func getValues(resourceMap map[string]*results.Resource) []*results.Resource {
	resources := make([]*results.Resource, 0, len(resourceMap))

	for _, resource := range resourceMap {
		resources = append(resources, resource)
	}

	return resources
}

func getAbsolutePath(file string, logger *gosteno.Logger) string {
	logger.Infof("Finding absolute path to $s", file)
	absolutePath, err := filepath.Abs(file)

	if err != nil {
		logger.Warnf("Error getting absolute path to $s using relative path due to %v", file, err)
		return file
	}

	return absolutePath
}
