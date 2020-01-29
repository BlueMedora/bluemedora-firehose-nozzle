// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package main

import (
	"flag"

	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/logger"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/nozzle"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/webserver"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/configuration"
)

const (
	defaultConfigLocation = "./config/bluemedora-firehose-nozzle.json"

	defaultLogDirectory = "./logs"
	nozzleLogFile       = "bm_nozzle.log"
	nozzleLogName       = "bm_firehose_nozzle"
	nozzleLogLevel      = "info"

	webserverLogFile = "bm_server.log"
	webserverLogName = "bm_server"
)

var (
	//Mode to run nozzle in. Webserver mode is for debugging purposes only
	runMode  = flag.String("mode", "normal", "Mode to run nozzle `normal` or `webserver`")
	logLevel = flag.String("log-level", nozzleLogLevel, "Set log level to control verbosity - defaults to info")
)

func main() {
	flag.Parse()

	if *runMode == "normal" {
		normalSetup()
	} else if *runMode == "webserver" {
		standUpWebServer()
	}
}

func normalSetup() {
	logger.CreateLogDirectory(defaultLogDirectory)

	l := logger.New(defaultLogDirectory, nozzleLogFile, nozzleLogName, *logLevel)
	l.Debug("working log")

	//Read in config
	c, err := configuration.New(defaultConfigLocation, l)
	if err != nil {
		l.Fatalf("Error parsing config file: %s", err.Error())
	}

	// Setup and start nozzle
	wsl := logger.New(defaultLogDirectory, webserverLogFile, webserverLogName, *logLevel)
	wsc := webserver.NewConfiguration(
		c.UAAUsername,
		c.UAAPassword,
		c.IdleTimeoutSeconds,
		c.MetricCacheDurationSeconds,
		c.WebServerPort,
		c.WebServerUseSSL,
		)

	ws := *webserver.New(wsc, wsl)

	nc := *nozzle.NewConfiguration(
    	c.UAAURL, 
    	c.UAAUsername, 
    	c.UAAPassword, 
    	c.TrafficControllerURL, 
    	c.SubscriptionID, 
    	c.DisableAccessControl, 
    	c.InsecureSSLSkipVerify)
    n := *nozzle.New(nc, l)
	n.Start()
	wsErrs := ws.Start()

	for {
		select {
		case envelope := <-n.Messages:
			l.Debug("printing a thing")
			ws.CacheEnvelnope(envelope)
		case err := <- wsErrs:
			// I think this becomes a fatal
			l.Fatalf("Error while running webserver: %s", err.Error())
		}
	}
}

func standUpWebServer() {
	logger := logger.New(defaultLogDirectory, webserverLogFile, webserverLogName, *logLevel)

	//Read in config
	config, err := configuration.New(defaultConfigLocation, logger)
	if err != nil {
		logger.Fatalf("Error parsing config file: %s", err.Error())
	}

	wsc := webserver.NewConfiguration(
		config.UAAUsername,
		config.UAAPassword,
		config.IdleTimeoutSeconds,
		config.MetricCacheDurationSeconds,
		config.WebServerPort,
		config.WebServerUseSSL,
		)
	server := webserver.New(wsc, logger)

	logger.Info("Starting webserver")
	errors := server.Start()

	select {
	case err := <-errors:
		logger.Fatalf("Error while running server: %s", err.Error())
	}
}
