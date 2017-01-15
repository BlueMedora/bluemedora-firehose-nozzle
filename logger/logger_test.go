// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package logger

import (
    "testing"
    "os"
    "fmt"
    "io/ioutil"
    "strings"
    "github.com/cloudfoundry/gosteno"
)

const (
    logDirectory = "../logs"
    logFile = "bm_nozzle.log"
    testLog = "test log"
    loggerName = "bm_firehose_nozzle"
    testLogLevel = "debug"
)

func TestLogDirectoryCreation(t *testing.T) {
    //Setup Envrionment
    setupEnvrionment(t)

    CreateLogDirectory(logDirectory)
    New(logDirectory, logFile, loggerName, testLogLevel)

    //See if log file was created
    checkLogFileExists(t)
}

func TestLogFileContents(t *testing.T) {
    //Setup Enivronment
    setupEnvrionment(t)

    CreateLogDirectory(logDirectory)
    logger := New(logDirectory, logFile, loggerName, testLogLevel)
    logger.Info(testLog)

    //Test if log contents contains test string
    checkLogContents(t)
}

func TestLogLevelOverride(t *testing.T) {
    //Setup Envrionment
    setupEnvrionment(t)
    //Override level with env var
    logLevel := "warn"
    os.Setenv(logLevelEnvVar, logLevel)

    CreateLogDirectory(logDirectory)
    New(logDirectory, logFile, loggerName, testLogLevel)

    //Test if log level is set from env
    checkLogLevel(t, logLevel)
}

func setupEnvrionment(t *testing.T) {
    t.Log("Removing logs directory...")
    if _, err := os.Stat(logDirectory); err != nil {
        os.RemoveAll(logDirectory)
    }
    t.Log("Removed logs directory")
}

func checkLogFileExists(t *testing.T) {
    t.Log("Check if log file bm_nozzle.log exists...")
    if _,err := os.Stat(fmt.Sprintf("%s/%s", logDirectory, logFile)); os.IsNotExist(err) {
        t.Fatalf("Log file bm_nozzle.log not created")
    }
}

func checkLogContents(t *testing.T) {
    fileBuffer, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", logDirectory, logFile))
    if err != nil {
        t.Fatalf("Failed to load log file: %s", err)
    }

    fileString := string(fileBuffer)

    t.Logf("Checking log contents... (expecting log contains: %s)", testLog)
    if !strings.Contains(fileString, testLog) {
        t.Errorf("Expected log contents to contain %s, string was not in log", testLog)
    }
}

func checkLogLevel(t *testing.T, logLevel string) {
    t.Log("Check if log level set from env")
    if l, _ := gosteno.GetLogLevel(logLevel); l != gosteno.LOG_WARN {
        t.Fatalf("Log level was not set to warn")
    }
}
