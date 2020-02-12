// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package main

import (
	"flag"
	"time"

	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/logger"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/nozzle"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/ttlcache"
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
	ws := *webserver.New(c, wsl)

    n := *nozzle.New(c, l)
   
	go func(ws webserver.WebServer, n nozzle.Nozzle){
	  n.Start()
      wsErrs := ws.Start()
      messages := n.Messages
      cache := ttlcache.GetInstance()
      
	  for {
		select {
		case m := <- messages:
			cache.UpdateResource(m)
		case err := <- wsErrs:
			// I think this becomes a fatal
			l.Fatalf("Error while running webserver: %s", err.Error())
		}
	  }
	}(ws, n) 
	
    // todo change to wait group
	for {
		time.Sleep(1*time.Hour)
	}
}

func standUpWebServer() {
	l := logger.New(defaultLogDirectory, webserverLogFile, webserverLogName, *logLevel)

	//Read in config
	c, err := configuration.New(defaultConfigLocation, l)
	if err != nil {
		l.Fatalf("Error parsing config file: %s", err.Error())
	}

	server := webserver.New(c, l)

	l.Info("Starting webserver")
	errors := server.Start()

	select {
	case err := <-errors:
		l.Fatalf("Error while running server: %s", err.Error())
	}
}