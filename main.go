// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package main

import (
	"flag"
	
	"github.com/BlueMedora/bluemedora-firehose-nozzle/bluemedorafirehosenozzle"
	"github.com/BlueMedora/bluemedora-firehose-nozzle/logger"
	"github.com/BlueMedora/bluemedora-firehose-nozzle/nozzleconfiguration"
	"github.com/BlueMedora/bluemedora-firehose-nozzle/webserver"
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
	runMode = flag.String("mode", "normal", "Mode to run nozzle `normal` or `webserver`")
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
    
	logger := logger.New(defaultLogDirectory, nozzleLogFile, nozzleLogName, *logLevel)
	logger.Debug("working log")

	//Read in config
	config, err := nozzleconfiguration.New(defaultConfigLocation, logger)
	if err != nil {
		logger.Fatalf("Error parsing config file: %s", err.Error())
	}

    //Setup and start nozzle
	server := createWebServer(config)

	nozzle := bluemedorafirehosenozzle.New(config, server, logger)
	nozzle.Start()

	if err != nil {
		logger.Fatalf("Error while running nozzle: %s", err.Error())
	}
}

func standUpWebServer() {
	logger := logger.New(defaultLogDirectory, webserverLogFile, webserverLogName, *logLevel)
    
    //Read in config
	config, err := nozzleconfiguration.New(defaultConfigLocation, logger)
	if err != nil {
		logger.Fatalf("Error parsing config file: %s", err.Error())
	}
    
    server := webserver.New(config, logger)
    
    logger.Info("Starting webserver")
    errors := server.Start(webserver.DefaultKeyLocation, webserver.DefaultCertLocation)
    
    select {
        case err := <-errors:
            logger.Fatalf("Error while running server: %s", err.Error())
    }
}

func createWebServer(config *nozzleconfiguration.NozzleConfiguration) *webserver.WebServer {
	logger := logger.New(defaultLogDirectory, webserverLogFile, webserverLogName, *logLevel)
	return webserver.New(config, logger)
}

