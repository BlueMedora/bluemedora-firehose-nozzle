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
	logLevel = flag.String("log-level", nozzleLogLevel, "Set log level to control verbosity - defaults to info")
)

func main() {
	flag.Parse()
	logger.CreateLogDirectory(defaultLogDirectory)
	l := logger.New(defaultLogDirectory, nozzleLogFile, nozzleLogName, *logLevel)

	cacheLogger := logger.New(defaultLogDirectory, cacheLogFile, cacheLogName, *logLevel)
	ttlcache.CreateInstance(cacheLogger)

	c, err := configuration.New(defaultConfigLocation, l)
	if err != nil {
		l.Fatalf("Error parsing config file: %s", err.Error())
	}

	wsl := logger.New(defaultLogDirectory, webserverLogFile, webserverLogName, *logLevel)
	ws := webserver.New(c, wsl)
	wsErrs := ws.Start()

	n := *nozzle.New(c, l)
	n.Start()

	cache := ttlcache.GetInstance()
	for {
		select {
		case m := <-n.Messages:
			cache.UpdateResource(m)
		case err := <-wsErrs:
			l.Fatalf("Error while running webserver: %s", err.Error())
		}
	}
}
