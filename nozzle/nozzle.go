// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package nozzle

import (
	// "crypto/tls"
	"fmt"
	// "time"

	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/loggregatorclient"

	"github.com/cloudfoundry/gosteno"
	"code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/cloudfoundry-incubator/uaago"
)

//BlueMedoraFirehoseNozzle consuems data from fire hose and exposes it via REST
type Nozzle struct {
	config     Configuration
	Messages   chan *loggregator_v2.Envelope
	logger     *gosteno.Logger
	client     loggregatorclient.Client
}

type Configuration struct {
	UAAURL                     string
	UAAUsername                string
	UAAPassword                string
	TrafficControllerURL       string
	SubscriptionID             string
	DisableAccessControl       bool
	InsecureSSLSkipVerify      bool
}

func NewConfiguration(uaaurl string, u string, p string, tcurl string, sub string, DisableAccessControl bool, skipVerify bool) *Configuration {
	return &Configuration{
		UAAURL: uaaurl,
		UAAUsername: u,
		UAAPassword: p,
		TrafficControllerURL: tcurl,
		SubscriptionID: sub,
		DisableAccessControl: DisableAccessControl,
		InsecureSSLSkipVerify: skipVerify,
	}
}

//New BlueMedoraFirhoseNozzle
func New(config Configuration, logger *gosteno.Logger) *Nozzle {
	return &Nozzle{
		config: config,
		logger: logger,
		Messages: make(chan *loggregator_v2.Envelope, 10000), //10k limit (evaluate)
		// ATP I dont like messages and errs not being passed in when we create the firehose
	}
}

//Start starts consuming events from firehose
func (n *Nozzle) Start() {
	n.logger.Info("Starting Blue Medora Firehose Nozzle")

	// var authToken string
	// if !nozzle.config.DisableAccessControl {
	// 	authToken = nozzle.fetchUAAAuthToken()
	// }
	authToken := n.fetchUAAAuthToken()

	// nozzle.logger.Debugf("Using auth token <%s>", authToken)

    // Z - TODO this shouldn't happen here, instead this should be handled by the main.go
    // main.go - n = bmfn.New
    // main.go - n.Start()
    // main.go - ws = ws.New(n.ttlcache)
    // main.go - err := ws.Start()
	// nozzle.serverErrs = nozzle.server.Start(webserver.DefaultKeyLocation, webserver.DefaultCertLocation)

	// nozzle.collectFromFirehose(authToken)
    n.client = *loggregatorclient.New(n.config.UAAURL, n.config.SubscriptionID, authToken, n.logger)	
    
    go func(messages chan *loggregator_v2.Envelope, es loggregator.EnvelopeStream){
    	// go forever
    	for {
    		for _, envelope := range es() {
    			messages <- envelope
    		}
    	}
    }(n.Messages, n.client.EnvelopeStream())

	n.logger.Info("Closing Blue Medora Firehose Nozzle")
}

func (n *Nozzle) fetchUAAAuthToken() string {
	n.logger.Debug("Fetching UAA authenticaiton token")

	UAAClient, err := uaago.NewClient(n.config.UAAURL)
	if err != nil {
		n.logger.Fatalf("Error creating UAA client: %s", err.Error())
	}

	var token string
	token, err = UAAClient.GetAuthToken(n.config.UAAUsername, n.config.UAAPassword, n.config.InsecureSSLSkipVerify)
	if err != nil {
		n.logger.Fatalf("Failed to get oauth token: %s.", err.Error())
	}

	n.logger.Debug(fmt.Sprintf("Successfully fetched UAA authentication token <%s>", token))
	return token
}


