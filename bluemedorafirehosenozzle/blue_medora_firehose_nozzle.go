// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package bluemedorafirehosenozzle

import (
	"crypto/tls"
	"fmt"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/ttlcache"
	"time"

	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/nozzleconfiguration"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/webserver"
	"github.com/cloudfoundry-incubator/uaago"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/noaa/consumer"
	noaaerrors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gorilla/websocket"
)

//BlueMedoraFirehoseNozzle consuems data from fire hose and exposes it via REST
type BlueMedoraFirehoseNozzle struct {
	config     *nozzleconfiguration.NozzleConfiguration
	errs       <-chan error
	messages   <-chan *events.Envelope
	serverErrs <-chan error
	logger     *gosteno.Logger
	consumer   *consumer.Consumer
	server     *webserver.WebServer
}

//New BlueMedoraFirhoseNozzle
func New(config *nozzleconfiguration.NozzleConfiguration, server *webserver.WebServer, logger *gosteno.Logger) *BlueMedoraFirehoseNozzle {
	return &BlueMedoraFirehoseNozzle{
		config: config,
		logger: logger,
		server: server,
	}
}

//Start starts consuming events from firehose
func (nozzle *BlueMedoraFirehoseNozzle) Start() error {
	nozzle.logger.Info("Starting Blue Medora Firehose Nozzle")

	var authToken string
	if !nozzle.config.DisableAccessControl {
		authToken = nozzle.fetchUAAAuthToken()
	}

	nozzle.logger.Debugf("Using auth token <%s>", authToken)

	nozzle.serverErrs = nozzle.server.Start(webserver.DefaultKeyLocation, webserver.DefaultCertLocation)

	nozzle.collectFromFirehose(authToken)
	err := nozzle.processMessages()

	nozzle.logger.Info("Closing Blue Medora Firehose Nozzle")
	return err
}

func (nozzle *BlueMedoraFirehoseNozzle) fetchUAAAuthToken() string {
	nozzle.logger.Debug("Fetching UAA authenticaiton token")

	UAAClient, err := uaago.NewClient(nozzle.config.UAAURL)
	if err != nil {
		nozzle.logger.Fatalf("Error creating UAA client: %s", err.Error())
	}

	var token string
	token, err = UAAClient.GetAuthToken(nozzle.config.UAAUsername, nozzle.config.UAAPassword, nozzle.config.InsecureSSLSkipVerify)
	if err != nil {
		nozzle.logger.Fatalf("Failed to get oauth token: %s.", err.Error())
	}

	nozzle.logger.Debug(fmt.Sprintf("Successfully fetched UAA authentication token <%s>", token))
	return token
}

func (nozzle *BlueMedoraFirehoseNozzle) collectFromFirehose(authToken string) {
	nozzle.consumer = consumer.New(nozzle.config.TrafficControllerURL, &tls.Config{InsecureSkipVerify: nozzle.config.InsecureSSLSkipVerify}, nil)

	debugPrinter := &BMDebugPrinter{nozzle.logger}
	nozzle.consumer.SetDebugPrinter(debugPrinter)
	nozzle.consumer.SetIdleTimeout(time.Duration(nozzle.config.IdleTimeoutSeconds) * time.Second)
	nozzle.messages, nozzle.errs = nozzle.consumer.Firehose(nozzle.config.SubscriptionID, authToken)
}

//Method blocks until error occurs
func (nozzle *BlueMedoraFirehoseNozzle) processMessages() error {
	//Set the TTLCache metric duration
	ttlcache.GetInstance().TTL = time.Duration(nozzle.config.MetricCacheDurationSeconds) * time.Second

	for {
		select {
		case envelope := <-nozzle.messages:
			nozzle.cacheEnvelope(envelope)
		case err := <-nozzle.serverErrs:
			if err != nil {
				nozzle.logger.Errorf("Error while running webserver: %s", err)
				return err
			}
		case err := <-nozzle.errs:
			if err != nil {
				nozzle.handleError(err)
				return err
			}
		}
	}
}

func (nozzle *BlueMedoraFirehoseNozzle) cacheEnvelope(envelope *events.Envelope) {
	nozzle.server.CacheEnvelope(envelope)
}

func (nozzle *BlueMedoraFirehoseNozzle) handleError(err error) {
	if retryErr, ok := err.(noaaerrors.RetryError); ok {
		err = retryErr.Err
	}

	switch closeError := err.(type) {
	case *websocket.CloseError:
		switch closeError.Code {
		case websocket.CloseNormalClosure:
			nozzle.logger.Infof("Connection closed normally: %s", err.Error())
		case websocket.ClosePolicyViolation:
			nozzle.logger.Errorf("Error while reading from firehose: %s", err.Error())
			nozzle.logger.Errorf("Disconnect due to nozzle not keeping up. Scale nozzle to prevent this problem.")
		default:
			nozzle.logger.Errorf("Error while reading from firehose: %s", err.Error())
		}
	default:
		nozzle.logger.Errorf("Error while reading from firehose: %s", err.Error())
	}

	nozzle.consumer.Close()
}
