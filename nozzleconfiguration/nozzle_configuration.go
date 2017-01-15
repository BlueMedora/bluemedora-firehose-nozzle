// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package nozzleconfiguration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cloudfoundry/gosteno"
)

const (
	uaaURLEnv                     = "BM_UAA_URL"
	uaaUsernameEnv                = "BM_UAA_USERNAME"
	uaaPasswordEnv                = "BM_UAA_PASSWORD"
	trafficControllerURLEnv       = "BM_TRAFFIC_CONTROLLER_URL"
	subscriptionIDEnv             = "BM_SUBSCRIPTION_ID"
	disableAccessControlEnv       = "BM_DISABLE_ACCESS_CONTROL"
	insecureSSLSkipVerifyEnv      = "BM_INSECURE_SSL_SKIP_VERIFY"
	idleTimeoutSecondsEnv         = "BM_IDLE_TIMEOUT_SECONDS"
	metricCacheDurationSecondsEnv = "BM_METRIC_CACHE_DURATION_SECONDS"
	webServerPortEnv              = "PORT"
	webServerUseSSLENV            = "BM_WEBSERVER_USE_SSL"
)

//NozzleConfiguration represents configuration file
type NozzleConfiguration struct {
	UAAURL                     string
	UAAUsername                string
	UAAPassword                string
	TrafficControllerURL       string
	SubscriptionID             string
	DisableAccessControl       bool
	InsecureSSLSkipVerify      bool
	IdleTimeoutSeconds         uint32
	MetricCacheDurationSeconds uint32
	WebServerPort              uint32
	WebServerUseSSL            bool
}

//New NozzleConfiguration
func New(configPath string, logger *gosteno.Logger) (*NozzleConfiguration, error) {
	configPath = getAbsolutePath(configPath, logger)

	configBuffer, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nil, fmt.Errorf("Unable to load config file bluemedora-firehose-nozzle.json: %s", err)
	}

	var nozzleConfig NozzleConfiguration
	err = json.Unmarshal(configBuffer, &nozzleConfig)
	if err != nil {
		return nil, fmt.Errorf("Error parsing config file bluemedora-firehose-nozzle.json: %s", err)
	}

	overrideWithEnvVar(uaaURLEnv, &nozzleConfig.UAAURL)
	overrideWithEnvVar(uaaUsernameEnv, &nozzleConfig.UAAUsername)
	overrideWithEnvVar(uaaPasswordEnv, &nozzleConfig.UAAPassword)
	overrideWithEnvVar(trafficControllerURLEnv, &nozzleConfig.TrafficControllerURL)
	overrideWithEnvVar(subscriptionIDEnv, &nozzleConfig.SubscriptionID)
	overrideWithEnvBool(disableAccessControlEnv, &nozzleConfig.DisableAccessControl)
	overrideWithEnvBool(insecureSSLSkipVerifyEnv, &nozzleConfig.InsecureSSLSkipVerify)
	overrideWithEnvUint32(idleTimeoutSecondsEnv, &nozzleConfig.IdleTimeoutSeconds)
	overrideWithEnvUint32(metricCacheDurationSecondsEnv, &nozzleConfig.MetricCacheDurationSeconds)
	overrideWithEnvUint32(webServerPortEnv, &nozzleConfig.WebServerPort)
	overrideWithEnvBool(webServerUseSSLENV, &nozzleConfig.WebServerUseSSL)

	logger.Debug(fmt.Sprintf("Loaded configuration to UAAURL <%s>, UAA Username <%s>, Traffic Controller URL <%s>, Disable Access Control <%v>, Insecure SSL Skip Verify <%v>",
		nozzleConfig.UAAURL, nozzleConfig.UAAUsername, nozzleConfig.TrafficControllerURL, nozzleConfig.DisableAccessControl, nozzleConfig.InsecureSSLSkipVerify))

	return &nozzleConfig, nil
}

func getAbsolutePath(configPath string, logger *gosteno.Logger) string {
	logger.Info("Finding absolute path to config file")
	absoluteConfigPath, err := filepath.Abs(configPath)

	if err != nil {
		logger.Warnf("Error getting absolute path to config file use relative path due to %v", err)
		return configPath
	}

	return absoluteConfigPath
}

func overrideWithEnvVar(name string, value *string) {
	envValue := os.Getenv(name)
	if envValue != "" {
		*value = envValue
	}
}

func overrideWithEnvUint32(name string, value *uint32) {
	envValue := os.Getenv(name)
	if envValue != "" {
		tmpValue, err := strconv.Atoi(envValue)
		if err != nil {
			panic(err)
		}
		*value = uint32(tmpValue)
	}
}

func overrideWithEnvBool(name string, value *bool) {
	envValue := os.Getenv(name)
	if envValue != "" {
		var err error
		*value, err = strconv.ParseBool(envValue)
		if err != nil {
			panic(err)
		}
	}
}
