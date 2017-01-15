// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package nozzleconfiguration

import (
    "testing"
    "os"
    "fmt"
    "encoding/json"
    "io/ioutil"
    "strings"
    "strconv"
    
    "github.com/BlueMedora/bluemedora-firehose-nozzle/logger"
) 

const (
    defaultLogDirectory = "../logs"
	nozzleLogFile       = "bm_nozzle.log"
	nozzleLogName       = "bm_firehose_nozzle"
    nozzleLogLevel      = "debug"

    configFile = "../config/bluemedora-firehose-nozzle.json"
    tempConfigFile = "../config/bluemedora-firehose-nozzle.json.real"
    
    testUAAURL = "UAAURL"
    testUsername = "username"
    testPassword = "password"
    testTrafficControllerURL = "traffic_url"
    testSubscriptionID = "bluemedora-nozzle"
    testDisableAccessControl = false
    testInsecureSSLSkipVerify = false
    testIdleTimeout = uint32(60)
    testMetricCacheDuration = uint32(60)
    testWebServerPort = uint32(8081)
    testWebServerUseSSL = true

    testEnvUAAURL = "env_UAAURL"
    testEnvUsername = "env_username"
    testEnvPassword = "env_password"
    testEnvTrafficControllerURL = "env_traffic_url"
    testEnvsubscriptionID = "env_bluemedora-nozzle"
    testEnvDisableAccessControl = "true"
    testEnvInsecureSSLSkipVerify = "true"
    testEnvIdleTimeout = "120"
    testEnvMetricCacheDuration = "90"
    testEnvWebServerPort = "9080"
    testEnvWebServerUseSSL = "true"
)

func TestConfigParsing(t *testing.T) {
    //Setup Environment
    err := setupGoodEnvironment(t)
    if err != nil {
        tearDownEnvironment(t)
        t.Fatalf("Setup failed due to: %s", err.Error())
    }
    
    t.Log("Creating configuration...")
    logger.CreateLogDirectory(defaultLogDirectory)
    logger := logger.New(defaultLogDirectory, nozzleLogFile, nozzleLogName, nozzleLogLevel)
    
    //Create new configuration
    var config *NozzleConfiguration
    config, err = New(configFile, logger)
    
    if err != nil {
        tearDownEnvironment(t)
        t.Fatalf("Error occrued while creating configuration %s", err)
    }
    
    //Test values
    t.Log(fmt.Sprintf("Checking UAA URL... (expected value: %s)", testUAAURL))
    if config.UAAURL != testUAAURL {
        t.Errorf("Expected UAA URL of %s, but received %s", testUAAURL, config.UAAURL)
    }
    
    t.Log(fmt.Sprintf("Checking UAA Username... (expected value: %s)", testUsername))
    if config.UAAUsername != testUsername {
        t.Errorf("Expected UAA Username of %s, but received %s", testUsername, config.UAAUsername)
    }

    t.Log(fmt.Sprintf("Checking UAA Password... (expected value: %s)", testPassword))
    if config.UAAPassword != testPassword {
        t.Errorf("Expected UAA Password of %s, but received %s", testPassword, config.UAAPassword)
    }
    
    t.Log(fmt.Sprintf("Checking Traffic Controller URL... (expected value: %s)", testTrafficControllerURL))
    if config.TrafficControllerURL != testTrafficControllerURL {
        t.Errorf("Expected Traffic Controller URL of %s, but received %s", testTrafficControllerURL, config.TrafficControllerURL)
    }

    t.Log(fmt.Sprintf("Checking Subscription ID... (expected value: %s)", testSubscriptionID))
    if config.SubscriptionID != testSubscriptionID {
        t.Errorf("Expected Subscription ID of %s, but received %s", testSubscriptionID, config.SubscriptionID)
    }
    
    t.Log(fmt.Sprintf("Checking Disable Access Control... (expected value: %v)", testDisableAccessControl))
    if config.DisableAccessControl != testDisableAccessControl {
        t.Errorf("Expected Disable Access Control of %v, but received %v", testDisableAccessControl, config.DisableAccessControl)
    }

    t.Log(fmt.Sprintf("Checking Insecure SSL Skip Verify... (expected value: %v)", testInsecureSSLSkipVerify))
    if config.InsecureSSLSkipVerify != testInsecureSSLSkipVerify {
        t.Errorf("Expected Insecure SSL Skip Verify of %v, but received %v", testInsecureSSLSkipVerify, config.InsecureSSLSkipVerify)
    }
    
    t.Log(fmt.Sprintf("Checking Idle Timeout... (expected value: %v)", testIdleTimeout))
    if config.IdleTimeoutSeconds != testIdleTimeout {
        t.Errorf("Expected Idle Timeout of %v, but received %v", testIdleTimeout, config.IdleTimeoutSeconds)
    }
    
    t.Log(fmt.Sprintf("Checking Metric Cache Duration... (expected value: %v)", testMetricCacheDuration))
    if config.MetricCacheDurationSeconds != testMetricCacheDuration {
        t.Errorf("Expected Metric Cache Duration of %v, but received %v", testMetricCacheDuration, config.MetricCacheDurationSeconds)
    }
    
    t.Log(fmt.Sprintf("Checking Web Server Port... (expected value: %v)", testWebServerPort))
    if config.WebServerPort != testWebServerPort {
        t.Errorf("Expected Web Server Port of %v, but received %v", testWebServerPort, config.WebServerPort)
    }

    t.Log(fmt.Sprintf("Checking Web Server Use SSL... (expected value: %v)", testWebServerUseSSL))
    if config.WebServerUseSSL != testWebServerUseSSL {
        t.Errorf("Expected Web Server Port of %v, but received %v", testWebServerUseSSL, config.WebServerPort)
    }
    
    err = tearDownEnvironment(t)
    if err != nil {
        t.Fatalf("Tear down failed due to: %s", err.Error())
    }
}

func TestBadConfigFile(t *testing.T) {
    err := setupBadEnvironment(t)
    if err != nil {
        tearDownEnvironment(t)
        t.Fatalf("Setup failed due to: %s", err.Error())
    }
    
    logger.CreateLogDirectory(defaultLogDirectory)
    logger := logger.New(defaultLogDirectory, nozzleLogFile, nozzleLogName, nozzleLogLevel)
    
    //Create new configuration
    t.Log("Checking loading of bad config file... (expecting error)")
    _, err = New(configFile, logger)
    
    if err != nil {
        if !strings.Contains(err.Error(), "Error parsing config file bluemedora-firehose-nozzle.json:") {
            t.Errorf("Expected error containing %s, but received %s", "Error parsing config file bluemedora-firehose-nozzle.json:", err.Error())
        }
    } else {
        t.Errorf("Expected error from loading a bad config file, but loaded correctly")
    }
    
    err = tearDownEnvironment(t)
    if err != nil {
        t.Fatalf("Tear down failed due to: %s", err.Error())
    }
}

func TestNoConfigFile(t *testing.T) {
    t.Log("Creating configuration...")
    logger.CreateLogDirectory(defaultLogDirectory)
    logger := logger.New(defaultLogDirectory, nozzleLogFile, nozzleLogName, nozzleLogLevel)
    
    //Create new configuration
    t.Log("Checking loading of non-existent file... (expecting error)")
    _, err := New("fake_file.json", logger)
    
    if err != nil {
        if !strings.Contains(err.Error(), "Unable to load config file bluemedora-firehose-nozzle.json:") {
            t.Errorf("Expected error containing %s, but received %s", "Unable to load config file bluemedora-firehose-nozzle.json:", err.Error())
        }
    } else {
        t.Errorf("Expected error from loading non-existsent file, but loaded correctly")
    }
}

func TestEnvironmentVariables(t *testing.T) {
    //Setup Environment
    err := setupGoodEnvironment(t)
    if err != nil {
        tearDownEnvironment(t)
        t.Fatalf("Setup failed due to: %s", err.Error())
    }
    
    t.Log("Creating configuration...")
    logger.CreateLogDirectory(defaultLogDirectory)
    logger := logger.New(defaultLogDirectory, nozzleLogFile, nozzleLogName, nozzleLogLevel)

    os.Setenv(uaaURLEnv, testEnvUAAURL)
    os.Setenv(uaaUsernameEnv, testEnvUsername)
    os.Setenv(uaaPasswordEnv, testEnvPassword)
    os.Setenv(trafficControllerURLEnv, testEnvTrafficControllerURL)
    os.Setenv(subscriptionIDEnv, testEnvsubscriptionID)
    os.Setenv(disableAccessControlEnv, testEnvDisableAccessControl)
    os.Setenv(insecureSSLSkipVerifyEnv, testEnvInsecureSSLSkipVerify)
    os.Setenv(idleTimeoutSecondsEnv, testEnvIdleTimeout)
    os.Setenv(metricCacheDurationSecondsEnv, testEnvMetricCacheDuration)
    os.Setenv(webServerPortEnv, testEnvWebServerPort)
    os.Setenv(webServerUseSSLENV, testEnvWebServerUseSSL)
    
    //Create new configuration
    var config *NozzleConfiguration
    config, err = New(configFile, logger)
    
    if err != nil {
        tearDownEnvironment(t)
        t.Fatalf("Error occrued while creating configuration %s", err)
    }

     //Test values
    t.Log(fmt.Sprintf("Checking UAA URL... (expected value: %s)", testEnvUAAURL))
    if config.UAAURL != testEnvUAAURL {
        t.Errorf("Expected UAA URL of %s, but received %s", testEnvUAAURL, config.UAAURL)
    }
    
    t.Log(fmt.Sprintf("Checking UAA Username... (expected value: %s)", testEnvUsername))
    if config.UAAUsername != testEnvUsername {
        t.Errorf("Expected UAA Username of %s, but received %s", testEnvUsername, config.UAAUsername)
    }

    t.Log(fmt.Sprintf("Checking UAA Password... (expected value: %s)", testEnvPassword))
    if config.UAAPassword != testEnvPassword {
        t.Errorf("Expected UAA Password of %s, but received %s", testEnvPassword, config.UAAPassword)
    }
    
    t.Log(fmt.Sprintf("Checking Traffic Controller URL... (expected value: %s)", testEnvTrafficControllerURL))
    if config.TrafficControllerURL != testEnvTrafficControllerURL {
        t.Errorf("Expected Traffic Controller URL of %s, but received %s", testEnvTrafficControllerURL, config.TrafficControllerURL)
    }

    t.Log(fmt.Sprintf("Checking Subscription ID... (expected value: %s)", testEnvsubscriptionID))
    if config.SubscriptionID != testEnvsubscriptionID {
        t.Errorf("Expected Subscription ID of %s, but received %s", testEnvsubscriptionID, config.SubscriptionID)
    }
    
    t.Log(fmt.Sprintf("Checking Disable Access Control... (expected value: %v)", testEnvDisableAccessControl))
    convertedDisableAccessControlValue, _ := strconv.ParseBool(testEnvDisableAccessControl) 
    if config.DisableAccessControl != convertedDisableAccessControlValue {
        t.Errorf("Expected Disable Access Control of %v, but received %v", testEnvDisableAccessControl, config.DisableAccessControl)
    }

    t.Log(fmt.Sprintf("Checking Insecure SSL Skip Verify... (expected value: %v)", testEnvInsecureSSLSkipVerify))
    convertedInsecureSSLSkipVerify, _ := strconv.ParseBool(testEnvInsecureSSLSkipVerify)
    if config.InsecureSSLSkipVerify != convertedInsecureSSLSkipVerify {
        t.Errorf("Expected Insecure SSL Skip Verify of %v, but received %v", testEnvInsecureSSLSkipVerify, config.InsecureSSLSkipVerify)
    }
    
    t.Log(fmt.Sprintf("Checking Idle Timeout... (expected value: %v)", testIdleTimeout))
    convertedtestEnvIdleTimeout, _ := strconv.Atoi(testEnvIdleTimeout)
    if config.IdleTimeoutSeconds != uint32(convertedtestEnvIdleTimeout) {
        t.Errorf("Expected Idle Timeout of %v, but received %v", testIdleTimeout, config.IdleTimeoutSeconds)
    }
    
    t.Log(fmt.Sprintf("Checking Metric Cache Duration... (expected value: %v)", testMetricCacheDuration))
    convertedtestEnvMetricCacheDuration, _ := strconv.Atoi(testEnvMetricCacheDuration)
    if config.MetricCacheDurationSeconds != uint32(convertedtestEnvMetricCacheDuration) {
        t.Errorf("Expected Metric Cache Duration of %v, but received %v", testMetricCacheDuration, config.MetricCacheDurationSeconds)
    }
    
    t.Log(fmt.Sprintf("Checking Web Server Port... (expected value: %v)", testWebServerPort))
    convertedtestEnvWebServerPort, _ := strconv.Atoi(testEnvWebServerPort)
    if config.WebServerPort != uint32(convertedtestEnvWebServerPort) {
        t.Errorf("Expected Web Server Port of %v, but received %v", testWebServerPort, config.WebServerPort)
    }

    t.Log(fmt.Sprintf("Checking Web Server Use SSL... (expected value: %v)", testEnvWebServerUseSSL))
    convertedtestEnvWebServerUseSSL, _ := strconv.ParseBool(testEnvWebServerUseSSL)
    if config.WebServerUseSSL != convertedtestEnvWebServerUseSSL {
        t.Errorf("Expected Web Server Port of %v, but received %v", testEnvWebServerUseSSL, config.WebServerPort)
    }
    
    err = tearDownEnvironment(t)
    if err != nil {
        t.Fatalf("Tear down failed due to: %s", err.Error())
    }
}

func setupGoodEnvironment(t *testing.T) error {
    t.Log("Setting up good environment...")
    
    err := renameConfigFile(t)
    if err != nil {
        return err
    }

    err = createGoodConfigFile(t)
    if err != nil {
        return err
    }
    
    t.Log("Setup good test environment")
    return nil
}

func setupBadEnvironment(t *testing.T) error {
    t.Log("Setting up bad envrionment...")
    
    err := renameConfigFile(t)
    if err != nil {
        return err
    }
    
    err = createBadConfigFile(t)
    if err != nil {
        return err
    }
    
    
    t.Log("Setup bad test envrionment")
    return nil
}

func renameConfigFile(t *testing.T) error {
    t.Log("Renaming real config file...")
    
    err := os.Rename(configFile, tempConfigFile)
    if err != nil {
        return fmt.Errorf("Error renaming config file. Ensure bluemedora-firehose-nozzle.json exists in config directory: %s", err)
    }
    
    t.Log("Renamed real config file")
    return nil
}

func createGoodConfigFile(t *testing.T) error {
    t.Log("Creating good config file...")
    
    message := NozzleConfiguration{
        testUAAURL, testUsername, 
        testPassword, testTrafficControllerURL, testSubscriptionID,
        testDisableAccessControl, testInsecureSSLSkipVerify, 
        testIdleTimeout, testMetricCacheDuration,
        testWebServerPort, testWebServerUseSSL}
        
    messageBytes, _ := json.Marshal(message)
    
    err := ioutil.WriteFile(configFile, messageBytes, os.ModePerm)
    if err != nil {
        return fmt.Errorf("Error creating good config file: %s", err)
    }
    
    t.Log("Created good config file")
    return nil
}

func createBadConfigFile(t *testing.T) error {
    t.Log("Creating bad config file...")
    
    _, err := os.Create(configFile)
    if err != nil {
        return fmt.Errorf("Error creating bad config file: %s", err)
    }
    
    t.Log("Created bad config file")
    return nil
}

func tearDownEnvironment(t *testing.T) error {
    t.Log("Tearing down test environment...")
    if _, err := os.Stat(tempConfigFile); os.IsNotExist(err) {
        t.Log("bluemedora-firehose-nozzle.json.real not found no clean up needed")
        return nil
    }
    
    if _, err := os.Stat(configFile); err == nil {
        err = os.Remove(configFile)
        if err != nil {
            return fmt.Errorf("Error removing test config file: %s", err)
        }
    }
    
    err := os.Rename(tempConfigFile, configFile)
    if err != nil {
        return fmt.Errorf("Error renaming config file. Ensure bluemedora-firehose-nozzle.json exists in config directory: %s", err)
    }
    
    t.Log("Tore down test environment")
    return nil
}