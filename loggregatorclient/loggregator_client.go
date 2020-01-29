// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package loggregatorclient

import (
	"context"
	"log"
    "net/http"
    "crypto/tls"

	"github.com/cloudfoundry/gosteno"
	"code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
)

type Client struct {
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

// "custom logger" for RLPGatewayClientLogger (:})
type RLPLogger struct {
	log *gosteno.Logger
}

// to impliment 'Log' and have messages passed to app logger
func (l RLPLogger) Write(p []byte) (n int, err error) {
	// Todo - no errors in processing are returned from envelopestream, but they are logged here.
	l.log.Debugf("RLP client: " + string(p))
	return len(p), nil
}

func New(address string, token string, subscriptionId string, logger *gosteno.Logger) *Client {
	c := loggregator.NewRLPGatewayClient(
		address,									
		loggregator.WithRLPGatewayClientLogger(log.New(&RLPLogger{log: logger}, "", log.LstdFlags)),
		loggregator.WithRLPGatewayHTTPClient(newRLPGatewayHttpClient(token, true)),
	)
	
	return &Client{
		RLPGatewayClient: c,
		stopConsumer: nil,
		shardId: subscriptionId,
	}
}

func (client *Client) EnvelopeStream() loggregator.EnvelopeStream {

	ctx, sfunc := context.WithCancel(context.Background())
	client.stopConsumer = sfunc
	es := client.RLPGatewayClient.Stream(
		ctx,
		&loggregator_v2.EgressBatchRequest{
			ShardId: client.shardId,
			Selectors: []*loggregator_v2.Selector{
				{
					Message: &loggregator_v2.Selector_Counter{
						Counter: &loggregator_v2.CounterSelector{},
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
	return es
}

func (c *Client) Stop() {
  c.stopConsumer()
}

func newRLPGatewayHttpClient(token string, noVerify bool) *rlpGatewayHttpClient {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: noVerify,
			},
		},
	}

	return &rlpGatewayHttpClient{
		token:        token,
		// disableACS:   disableACS,
		// tokenFetcher: tokenFetcher,
		client:       client,
	}
}

func (rlp *rlpGatewayHttpClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", rlp.token)

	// TODO: handle bad token and retry getting another one (might depend on access control)
	resp, err := rlp.client.Do(req)
	return resp, err
}
