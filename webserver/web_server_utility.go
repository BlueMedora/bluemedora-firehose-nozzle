// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package webserver

import (
	"fmt"
	"path/filepath"

	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/sonde-go/events"
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

func createEnvelopeKey(envelope *events.Envelope) string {
	return fmt.Sprintf("%s | %s | %s | %s", envelope.GetDeployment(), envelope.GetJob(), envelope.GetIndex(), envelope.GetIp())
}

func addMetric(envelope *events.Envelope, valueMetricMap map[string]float64, counterMetricMap map[string]float64, logger *gosteno.Logger) {
	if envelope.GetEventType() == events.Envelope_ValueMetric {
		valueMetric := envelope.GetValueMetric()

		valueMetricMap[valueMetric.GetName()] = valueMetric.GetValue()
		logger.Debugf("Adding Counter Event Name %s, Value %d", valueMetric.GetName(), valueMetric.GetValue())
	} else if envelope.GetEventType() == events.Envelope_CounterEvent {
		counterEvent := envelope.GetCounterEvent()

		counterMetricMap[counterEvent.GetName()] = float64(counterEvent.GetTotal())
		logger.Debugf("Adding Counter Event Name %s, Value %d", counterEvent.GetName(), counterEvent.GetTotal())
	} else {
		logger.Warnf("Unkown event type %s", envelope.GetEventType())
	}
}

func getValues(resourceMap map[string]Resource) []Resource {
	resources := make([]Resource, 0, len(resourceMap))

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
