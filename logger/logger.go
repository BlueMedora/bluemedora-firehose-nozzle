// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package logger


import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cloudfoundry/gosteno"
)

const (
	stdOutLogging  = "BM_STDOUT_LOGGING"
	logLevelEnvVar = "BM_LOG_LEVEL"
)

//New logger
func New(logDirectory string, logFile string, loggerName string, logLevel string) *gosteno.Logger {
	loggingConfig := &gosteno.Config{
		Sinks:		 make([]gosteno.Sink, 1),
		Level:     computeLevel(logLevel),
		Codec:     gosteno.NewJsonCodec(),
		EnableLOC: true,
	}


	//todo check value of env var
	if envValue := os.Getenv(stdOutLogging); envValue != "" {
		value, err := strconv.ParseBool(envValue)

		if err != nil {
			log.Fatalf("Failed to read in environment variable %s due to %v", stdOutLogging, err)
		}

		if (value) {
			log.Print("Logging to stdout")
			loggingConfig.Sinks[0] = gosteno.NewIOSink(os.Stdout)
		} else {
			log.Print("Logging to file")
			loggingConfig.Sinks[0] = gosteno.NewFileSink(fmt.Sprintf("%s/%s", getAbsolutePath(logDirectory), logFile))
		}
	} else {
		log.Print("Logging to file")
		loggingConfig.Sinks[0] = gosteno.NewFileSink(fmt.Sprintf("%s/%s", getAbsolutePath(logDirectory), logFile))
	}

	gosteno.Init(loggingConfig)
	return gosteno.NewLogger(loggerName)
}

//CreateLogDirectory clears out old directory and creates a new one
func CreateLogDirectory(logDirectory string) {
	absoluteDirectoryPath := getAbsolutePath(logDirectory)

	log.Printf("Using path %s", absoluteDirectoryPath)

	if _, err := os.Stat(fmt.Sprintf("%s/", absoluteDirectoryPath)); err == nil {
		os.RemoveAll(fmt.Sprintf("%s", absoluteDirectoryPath))
	}

	os.MkdirAll(fmt.Sprintf("%s/", absoluteDirectoryPath), os.ModePerm)
}

func getAbsolutePath(logDirectory string) string {
    log.Print("Finding absolute path to log directory")
	absolutelogDirectory, err := filepath.Abs(logDirectory)

	if err != nil {
        log.Printf("Error getting absolute path to log directory using relative path due to %v", err)
		return logDirectory
	}

	return absolutelogDirectory
}

func computeLevel(name string) gosteno.LogLevel {
	if envValue := os.Getenv(logLevelEnvVar); envValue != "" {
		name = envValue
	}
	ll, err := gosteno.GetLogLevel(name)
	if err != nil {
		return gosteno.LOG_INFO
	}
	return ll
}
