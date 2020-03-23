// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package configuration

import (
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry/gosteno"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

const (
	uaaURLEnv                     = "UAA_HOST"
	uaaUsernameEnv                = "BM_UAA_USERNAME"
	uaaPasswordEnv                = "BM_UAA_PASSWORD"
	cloudControllerURLEnv         = "CC_HOST"
	rlpUrlEnv                     = "RLP_URL"
	subscriptionIDEnv             = "BM_SUBSCRIPTION_ID"
	disableAccessControlEnv       = "BM_DISABLE_ACCESS_CONTROL"
	insecureSSLSkipVerifyEnv      = "BM_INSECURE_SSL_SKIP_VERIFY"
	idleTimeoutSecondsEnv         = "BM_IDLE_TIMEOUT_SECONDS"
	metricCacheDurationSecondsEnv = "BM_METRIC_CACHE_DURATION_SECONDS"
	webServerPortEnv              = "PORT"
	webServerUseSSLENV            = "BM_WEBSERVER_USE_SSL"
	webServerCertLocation         = "BM_WEBSERVER_CERT_LOCATION"
	webServerKeyLocation          = "BM_WEBSERVER_KEY_LOCATION"
)

//NozzleConfiguration represents configuration file
type Configuration struct {
	UAAURL                     string
	UAAUsername                string
	UAAPassword                string
	RLPURL                     string
	SubscriptionID             string
	DisableAccessControl       bool
	InsecureSSLSkipVerify      bool
	IdleTimeoutSeconds         uint32
	MetricCacheDurationSeconds uint32
	WebServerPort              uint32
	WebServerUseSSL            bool
	WebServerCertLocation      string
	WebServerKeyLocation       string
}

//New NozzleConfiguration
func New(configPath string, logger *gosteno.Logger) (*Configuration, error) {
	configPath = getAbsolutePath(configPath, logger)

	configBuffer, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nil, fmt.Errorf("Unable to load config file bluemedora-firehose-nozzle.json: %s", err)
	}

	var c Configuration
	err = json.Unmarshal(configBuffer, &c)
	if err != nil {
		return nil, fmt.Errorf("Error parsing config file bluemedora-firehose-nozzle.json: %s", err)
	}

	overrideWithEnvVar(uaaURLEnv, &c.UAAURL)
	overrideWithEnvVar(uaaUsernameEnv, &c.UAAUsername)
	overrideWithEnvVar(uaaPasswordEnv, &c.UAAPassword)

	overrideWithEnvVar(subscriptionIDEnv, &c.SubscriptionID)
	overrideWithEnvBool(disableAccessControlEnv, &c.DisableAccessControl)
	overrideWithEnvBool(insecureSSLSkipVerifyEnv, &c.InsecureSSLSkipVerify)
	overrideWithEnvUint32(idleTimeoutSecondsEnv, &c.IdleTimeoutSeconds)
	overrideWithEnvUint32(metricCacheDurationSecondsEnv, &c.MetricCacheDurationSeconds)
	overrideWithEnvUint32(webServerPortEnv, &c.WebServerPort)
	overrideWithEnvBool(webServerUseSSLENV, &c.WebServerUseSSL)
	overrideWithEnvVar(webServerCertLocation, &c.WebServerCertLocation)
	overrideWithEnvVar(webServerKeyLocation, &c.WebServerKeyLocation)

	// we use the specified RLP URL over converting the CC URL
	rlp := os.Getenv(rlpUrlEnv)
	if rlp != "" {
		c.RLPURL = rlp
	} else {
		overrideWithEnvVar(cloudControllerURLEnv, &c.RLPURL)
		r := regexp.MustCompile("://(api)")
		c.RLPURL = r.ReplaceAllString(c.RLPURL, "://log-stream")
	}

	logger.Debug(fmt.Sprintf("Loaded configuration to UAAURL <%s>, UAA Username <%s>, RLP URL <%s>, Disable Access Control <%v>, Insecure SSL Skip Verify <%v>",
		c.UAAURL, c.UAAUsername, c.RLPURL, c.DisableAccessControl, c.InsecureSSLSkipVerify))

	return &c, nil
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
