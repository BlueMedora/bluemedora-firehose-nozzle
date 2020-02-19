// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package nozzle

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/configuration"

	"code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/cloudfoundry-incubator/uaago"
	"github.com/cloudfoundry/gosteno"
)

type Nozzle struct {
	client   *loggregator.RLPGatewayClient
	config   *configuration.Configuration
	logger   *gosteno.Logger
	Messages chan *loggregator_v2.Envelope
}

func New(config *configuration.Configuration, logger *gosteno.Logger) *Nozzle {
	l := log.New(NewRLPLogger(logger), "", log.LstdFlags)

	c := loggregator.NewRLPGatewayClient(
		config.TrafficControllerURL,
		loggregator.WithRLPGatewayClientLogger(l),
		loggregator.WithRLPGatewayHTTPClient(newNozzleHTTPClient(config, logger)),
	)

	return &Nozzle{
		client:   c,
		config:   config,
		logger:   logger,
		Messages: make(chan *loggregator_v2.Envelope, 10000), //max 10k unprocessed envelopes. This is a reasonable limit we should be able to keep up with.
	}
}

// Start starts consuming events from firehose
func (n *Nozzle) Start() {
	n.logger.Info("Starting Blue Medora Firehose Nozzle")

	go func() {
		es := n.envelopeStream()
		for {
			for _, e := range es() {
				n.Messages <- e
			}
		}
	}()
}

func (n *Nozzle) envelopeStream() loggregator.EnvelopeStream {
	ctx := context.Background()
	return n.client.Stream(
		ctx,
		&loggregator_v2.EgressBatchRequest{
			ShardId: n.config.SubscriptionID,
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
}

type RLPLogger struct {
	log *gosteno.Logger
}

func NewRLPLogger(logger *gosteno.Logger) *RLPLogger {
	return &RLPLogger{
		log: logger,
	}
}

func (l RLPLogger) Write(p []byte) (n int, err error) {
	l.log.Info("RLP client: " + string(p))
	return len(p), nil
}

type nozzleHTTPClient struct {
	client *http.Client
	config *configuration.Configuration
	logger *gosteno.Logger
	token  string
}

func newNozzleHTTPClient(config *configuration.Configuration, logger *gosteno.Logger) *nozzleHTTPClient {
	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.InsecureSSLSkipVerify,
			},
		},
	}

	return &nozzleHTTPClient{
		client: c,
		config: config,
		logger: logger,
	}
}

func (c *nozzleHTTPClient) fetchToken() string {
	c.logger.Debug("Fetching UAA authenticaiton token")

	authClient, uaaErr := uaago.NewClient(c.config.UAAURL)
	if uaaErr != nil {
		c.logger.Fatalf("Error creating UAA client: %s", uaaErr.Error())
	}

	t, err := authClient.GetAuthToken(c.config.UAAUsername, c.config.UAAPassword, c.config.InsecureSSLSkipVerify)
	if err != nil {
		c.logger.Fatalf("Failed to get oauth token: %s.", err.Error())
	}

	c.logger.Debugf("Successfully fetched UAA authentication token <%s>", t)
	return t
}

func (c *nozzleHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if !c.config.DisableAccessControl && c.token == "" {
		c.token = c.fetchToken()
	}

	req.Header.Set("Authorization", c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Fatalf("Failure to make http request. %s", err.Error())
	}

	if !c.config.DisableAccessControl && (resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden) {
		time.Sleep(10 * time.Millisecond) // todo figure out why
		c.token = c.fetchToken()
		req.Header.Set("Authorization", c.token)
		resp, err = c.client.Do(req)
	}

	return resp, err
}
