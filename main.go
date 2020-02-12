// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package main

import (
	"flag"

	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/configuration"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/logger"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/nozzle"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/ttlcache"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/webserver"

	"github.com/cloudfoundry/gosteno"
)

const (
	defaultConfigLocation = "./config/bluemedora-firehose-nozzle.json"

	defaultLogDirectory = "./logs"
	nozzleLogFile       = "bm_nozzle.log"
	nozzleLogName       = "bm_firehose_nozzle"
	nozzleLogLevel      = "info"

	webserverLogFile = "bm_server.log"
	webserverLogName = "bm_server"

	cacheLogFile = "bm_cache.log"
	cacheLogName = "bm_cache"
)

var (
	//Mode to run nozzle in. Webserver mode is for debugging purposes only
	runMode  = flag.String("mode", "normal", "Mode to run nozzle `normal` or `webserver`")
	logLevel = flag.String("log-level", nozzleLogLevel, "Set log level to control verbosity - defaults to info")
)

func main() {
	flag.Parse()
	logger.CreateLogDirectory(defaultLogDirectory)
	l := logger.New(defaultLogDirectory, nozzleLogFile, nozzleLogName, *logLevel)

	cacheLogger := logger.New(defaultLogDirectory, cacheLogFile, cacheLogName, *logLevel)
	ttlcache.CreateInstance(cacheLogger)

	//Read in config
	c, err := configuration.New(defaultConfigLocation, l)
	if err != nil {
		l.Fatalf("Error parsing config file: %s", err.Error())
	}

	if *runMode == "normal" {
		normalSetup(c, l)
	} else {
		webServerSetup(c)
	}
}

func normalSetup(config *configuration.Configuration, nozzleLogger *gosteno.Logger) nozzle.Nozzle {
	// Setup and start webserver
	wsl := logger.New(defaultLogDirectory, webserverLogFile, webserverLogName, *logLevel)
	ws := webserver.New(config, wsl)
	wsErrs := ws.Start()

	// Setup and start nozzle
	n := *nozzle.New(config, nozzleLogger)
	n.Start()

	cache := ttlcache.GetInstance()
	for {
		select {
		case m := <-n.Messages:
			cache.UpdateResource(m)
		case err := <-wsErrs:
			nozzleLogger.Fatalf("Error while running webserver: %s", err.Error())
		}
	}
	return n
}

func webServerSetup(config *configuration.Configuration) <-chan error {
	l := logger.New(defaultLogDirectory, webserverLogFile, webserverLogName, *logLevel)

	ws := webserver.New(config, l)
	wsErrs := ws.Start()

	l.Info("Starting webserver")
	for {
		select {
		case err := <-wsErrs:
			l.Fatalf("Error while running server: %s", err.Error())
		}
	}
}
