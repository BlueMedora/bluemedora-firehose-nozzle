// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package loggregatorclient

import (
	"context"
	"log"
    "os"
    "net/http"
	


	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/nozzleconfiguration"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/webserver"
	"github.com/cloudfoundry-incubator/uaago"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gorilla/websocket"

	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/go-loggregator"
	"github.com/cloudfoundry/go-loggregator/rpc/loggregator_v2"
)

type LoggregatorClient struct {
	RLPGatewayClient *loggregator.RLPGatewayClient
	stopConsumer context.CancelFunc
	shardId string
}

type rlpGatewayHttpClient struct {
	token        string
	// disableACS   bool
	//tokenFetcher AuthTokenFetcher
	client       *http.Client
}

func newRLPGatewayHttpClient(token String, noVerify bool) *rlpGatewayClientDoer {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: noVerify,
			},
		},
	}

	return &rlpGatewayClientDoer{
		token:        token,
		// disableACS:   disableACS,
		// tokenFetcher: tokenFetcher,
		client:       client,
	}
}

func (rlp *newRLPGatewayHttpClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", rlp.token)

	// TODO: handle bad token and retry getting another one (might depend on access control)
	resp, err := rlp.client.Do(req)
	return resp, err
}

// "custom logger" for RLPGatewayClientLogger (:})
type RLPLogger struct {
	log *gosteno.Logger
}

func New(address String, token String, subscriptionId String, logger *gosteno.Logger) *LoggregatorClient {
	c := loggregator.NewRLPGatewayClient(
		address,									
		loggregator.WithRLPGatewayClientLogger(&RLPLogger{log: logger}, "", log.LstdFlags)),
		loggregator.WithRLPGatewayHTTPClient(newRLPGatewayHttpClient(token, true)),
	)
	
	return &LoggregatorClient{
		rlpClient: c,
		stopConsumer: nil,
		shardId: subscriptionId,
	}
}

func (client *LoggregatorClient) EnvelopeStream() loggregator.EnvelopeStream {
	ctx, client.stopConsumer = ctx.WithCancel(context.Background())
	eventStream := client.RLPGatewayClient.Stream(
		ctx,
		&loggregator_v2.EgressBatchRequest{
			ShardId: client.shardId,
			Selectors: []*loggregator_v2.Selector{
				{
					Message: &loggregator_v2.Selector_Counter{
						Counter: &loggregator_v2.CounterSelector{}
					},
				},
				{
					Message: &loggregator_v2.Selector_Gauge{
						Gauge: &loggregator_v2.GaugeSelector{},
					},
				},
			},
		},
	)
	return eventStream
}

func (c *LoggregatorClient) Stop() {
  c.stopConsumer()
}

