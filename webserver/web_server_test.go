// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package webserver

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/logger"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/nozzleconfiguration"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/testhelpers"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/ttlcache"
	"github.com/cloudfoundry/sonde-go/events"
)

const (
	defaultConfigLocation = "../config/bluemedora-firehose-nozzle.json"

	defaultLogDirectory = "../logs"
	webserverLogFile    = "bm_server.log"
	webserverLogName    = "bm_server"
	webserverLogLevel   = "debug"

	testCertLocation = "../certs/cert.pem"
	testKeyLocation  = "../certs/key.pem"
)

var (
	server *WebServer
	config *nozzleconfiguration.NozzleConfiguration
)

func TestTokenEndpoint(t *testing.T) {
	server, config = createWebServer(t)

	t.Log("Setting up server environment...")
	testhelpers.GenerateCertFiles()
	errors := server.Start(testKeyLocation, testCertLocation)

	//Handle errors from server
	go func() {
		select {
		case err := <-errors:
			if err != nil {
				t.Fatalf("Error with server: %s", err.Error())
			}
		}
	}()

	t.Log("Waiting a minute to allow total setup of webserver on travis")
	time.Sleep(time.Duration(1) * time.Minute)

	client := createHTTPClient(t)

	//Token tests
	tokenEndPointTest(t, client, config)
}

func TestNoTokenEndpointRequest(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//No Token tests
	noTokenEndPointTest(t, client, config.WebServerPort, server)
}

func TestPutRequestToResourceEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Put request to resource endpoint test
	resourcePutEndPointTest(t, client, config.WebServerPort)
}

func TestNoCacheDataEndpointRequest(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	//Cleared cache test
	noCachedDataTest(t, client, token, config.WebServerPort, server)
}

func TestMetronAgentEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, metronAgentOrigin, "metron_agents", server)
}

func TestSyslogDrainBinderEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, syslogDrainBinderOrigin, "syslog_drains", server)
}

func TestTPSWatcherEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, tpsWatcherOrigin, "tps_watchers", server)
}

func TestTPSListenerEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, tpsListenerOrigin, "tps_listeners", server)
}

func TestStagerEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, stagerOrigin, "stagers", server)
}

func TestSSHProxiesEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, sshProxyOrigin, "ssh_proxies", server)
}

func TestSenderEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, senderOrigin, "senders", server)
}

func TestRouteEmitterEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, routeEmitterOrigin, "route_emitters", server)
}

func TestRepEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, repOrigin, "reps", server)
}

func TestReceptorEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, receptorOrigin, "receptors", server)
}

func TestNSYNCListenerEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, nsyncListenerOrigin, "nsync_listeners", server)
}

func TestNSYNCBulkerEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, nsyncBulkerOrigin, "nsync_bulkers", server)
}

func TestGardenLinuxEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, gardenLinuxOrigin, "garden_linuxs", server)
}

func TestFileServerEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, fileServerOrigin, "file_servers", server)
}

func TestFetcherEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, fetcherOrigin, "fetchers", server)
}

func TestConvergerEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, convergerOrigin, "convergers", server)
}

func TestCCUploaderEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, ccUploaderOrigin, "cc_uploaders", server)
}

func TestbbsEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, bbsOrigin, "bbs", server)
}

func TestAuctioneerEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, auctioneerOrigin, "auctioneers", server)
}

func TestetcdEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, etcdOrigin, "etcds", server)
}

func TestDopplerServerEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, dopplerServerOrigin, "doppler_servers", server)
}

func TestCloudControllerEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, cloudControllerOrigin, "cloud_controllers", server)
}

func TestTrafficControllerEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, trafficControllerOrigin, "traffic_controllers", server)
}

func TestGoRouterEndpoint(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	endPointTest(t, client, token, config.WebServerPort, goRouterOrigin, "gorouters", server)
}

func TestTokenTimeout(t *testing.T) {
	if server == nil {
		t.Fatalf("Server failed to initialize in first test")
	}

	client := createHTTPClient(t)

	//Retrieve token for other endpoint test
	token := getToken(t, client, config)

	t.Log("Waiting 3 minutes to enusre token invalidates")
	time.Sleep(time.Duration(3) * time.Minute)

	request := createResourceRequest(t, token, config.WebServerPort, "gorouters")

	t.Logf("Check if server response to invalid token usage... (expecting status code: %v)", http.StatusUnauthorized)
	response, err := client.Do(request)

	if err != nil {
		t.Errorf("Error occured while hitting endpoint: %s", err.Error())
	} else if response.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expecting status code %v, but received %v", http.StatusUnauthorized, response.StatusCode)
	}
}

/** Tests **/
func tokenEndPointTest(t *testing.T, client *http.Client, config *nozzleconfiguration.NozzleConfiguration) {
	t.Log("Running token request tests...")
	badCredentialTokenTest(t, client, config)
	noCredentialTokenTest(t, client, config)
	goodTokenRequestTest(t, client, config)
	putTokenRequestTest(t, client, config)
	t.Log("Finished token request tests")
}

func goodTokenRequestTest(t *testing.T, client *http.Client, config *nozzleconfiguration.NozzleConfiguration) {
	tokenRequest := createTokenRequest("GET", config.UAAUsername, config.UAAPassword, config.WebServerPort, t)

	t.Logf("Check if server responses to good token request... (expecting status code: %v)", http.StatusOK)
	response, err := client.Do(tokenRequest)
	if err != nil {
		t.Fatalf("Error occured while requesting token: %s", err.Error())
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expecting status code %v, but received %v", http.StatusOK, response.StatusCode)
	}
}

func badCredentialTokenTest(t *testing.T, client *http.Client, config *nozzleconfiguration.NozzleConfiguration) {
	tokenRequest := createTokenRequest("GET", "baduser", "badPass", config.WebServerPort, t)

	t.Logf("Check if server responses to a bad credential token request... (expecting status code: %v)", http.StatusUnauthorized)
	response, err := client.Do(tokenRequest)
	if err != nil {
		t.Fatalf("Error occured while requesting token: %s", err.Error())
	}

	if response.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expecting status code %v, but received %v", http.StatusUnauthorized, response.StatusCode)
	}
}

func noCredentialTokenTest(t *testing.T, client *http.Client, config *nozzleconfiguration.NozzleConfiguration) {
	tokenRequest := createTokenRequest("GET", "", "", config.WebServerPort, t)

	t.Logf("Check if server responses to a no credential token request... (expecting status code: %v)", http.StatusBadRequest)
	response, err := client.Do(tokenRequest)
	if err != nil {
		t.Fatalf("Error occured while requesting token: %s", err.Error())
	}

	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("Expecting status code %v, but received %v", http.StatusBadRequest, response.StatusCode)
	}
}

func putTokenRequestTest(t *testing.T, client *http.Client, config *nozzleconfiguration.NozzleConfiguration) {
	tokenRequest := createTokenRequest("PUT", config.UAAUsername, config.UAAPassword, config.WebServerPort, t)

	t.Logf("Check if server responses to put token request... (expecting status code: %v)", http.StatusMethodNotAllowed)
	response, err := client.Do(tokenRequest)
	if err != nil {
		t.Fatalf("Error occured while requesting token: %s", err.Error())
	}

	if response.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expecting status code %v, but received %v", http.StatusMethodNotAllowed, response.StatusCode)
	}
}

func endPointTest(t *testing.T, client *http.Client, token string, port uint32, endPointOrigin string, endPointString string, server *WebServer) {
	cacheEnvelope(endPointOrigin, server)

	request := createResourceRequest(t, token, port, endPointString)

	t.Logf("Check if server response to valid /%s request... (expecting status code: %v)", endPointString, http.StatusOK)
	response, err := client.Do(request)

	if err != nil {
		t.Errorf("Error occured while hitting endpoint: %s", err.Error())
	} else if response.StatusCode != http.StatusOK {
		t.Errorf("Expecting status code %v, but received %v", http.StatusOK, response.StatusCode)
	}
}

func noTokenEndPointTest(t *testing.T, client *http.Client, port uint32, server *WebServer) {
	request := createResourceRequest(t, "", port, "gorouters")

	t.Logf("Check if server response to no token request... (expecting status code: %v)", http.StatusUnauthorized)
	response, err := client.Do(request)

	if err != nil {
		t.Errorf("Error occured while hitting endpoint: %s", err.Error())
	} else if response.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expecting status code %v, but received %v", http.StatusUnauthorized, response.StatusCode)
	}
}

func resourcePutEndPointTest(t *testing.T, client *http.Client, port uint32) {
	request, _ := http.NewRequest("PUT", fmt.Sprintf("https://localhost:%d/%s", port, "gorouters"), nil)

	t.Logf("Check if server response to put resource endpoint request... (expecting status code: %v)", http.StatusMethodNotAllowed)
	response, err := client.Do(request)

	if err != nil {
		t.Errorf("Error occured while hitting endpoint: %s", err.Error())
	} else if response.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expecting status code %v, but received %v", http.StatusMethodNotAllowed, response.StatusCode)
	}
}

func noCachedDataTest(t *testing.T, client *http.Client, token string, port uint32, server *WebServer) {
	cacheEnvelope(goRouterOrigin, server)

	request := createResourceRequest(t, token, port, "gorouters")

	//Give the cache time to clearout
	time.Sleep(20 * time.Second)

	t.Logf("Check if server response to put resource endpoint request... (expecting status code: %v)", http.StatusNoContent)
	response, err := client.Do(request)

	if err != nil {
		t.Errorf("Error occured while hitting endpoint: %s", err.Error())
	} else if response.StatusCode != http.StatusNoContent {
		t.Errorf("Expecting status code %v, but received %v", http.StatusNoContent, response.StatusCode)
	}
}

/** Utility Functions **/
func createWebServer(t *testing.T) (*WebServer, *nozzleconfiguration.NozzleConfiguration) {
	t.Log("Creating webserver...")
	logger.CreateLogDirectory(defaultLogDirectory)
	logger := logger.New(defaultLogDirectory, webserverLogFile, webserverLogName, webserverLogLevel)

	config, err := nozzleconfiguration.New(defaultConfigLocation, logger)
	if err != nil {
		t.Fatalf("Error while loading configuration: %s", err.Error())
	}

	cache := ttlcache.GetInstance()
	cache.TTL = time.Second

	t.Log("Created webserver")
	return New(config, logger), config
}

func createHTTPClient(t *testing.T) *http.Client {
	t.Log("Creating a test client...")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	t.Log("Created a test client")
	return &http.Client{Transport: tr}
}

func createTokenRequest(httpmethod string, username string, password string, port uint32, t *testing.T) *http.Request {
	t.Log("Creating token request...")
	request, err := http.NewRequest(httpmethod, fmt.Sprintf("https://localhost:%d/token", port), nil)
	if err != nil {
		t.Fatalf("Error creating token request: %s", err.Error())
	}

	if username != "" || password != "" {
		request.Header.Add(headerUsernameKey, username)
		request.Header.Add(headerPasswordKey, password)
	}

	t.Log("Created token request")
	return request
}

func getToken(t *testing.T, client *http.Client, config *nozzleconfiguration.NozzleConfiguration) string {
	t.Log("Requesting token...")
	tokenRequest := createTokenRequest("GET", config.UAAUsername, config.UAAPassword, config.WebServerPort, t)
	response, err := client.Do(tokenRequest)

	if err != nil {
		t.Fatalf("Error occured while requsting token: %s", err.Error())
	}

	t.Log("Requested token")
	return response.Header.Get(headerTokenKey)
}

func createResourceRequest(t *testing.T, token string, port uint32, endpoint string) *http.Request {
	t.Logf("Creating resource request...")
	request, err := http.NewRequest("GET", fmt.Sprintf("https://localhost:%d/%s", port, endpoint), nil)

	if err != nil {
		t.Fatalf("Error occured while formatting request %s: %s", endpoint, err.Error())
	}

	request.Header.Add(headerTokenKey, token)

	t.Logf("Created resource request")
	return request
}

func cacheEnvelope(originType string, server *WebServer) {
	deployment := "deployment"
	eventType := events.Envelope_ValueMetric
	job := "job"
	index := "0"
	ip := "127.0.0.1"
	metricName := "metric"
	value := float64(100)
	unit := "unit"

	valueMetric := events.ValueMetric{
		Name:  &metricName,
		Value: &value,
		Unit:  &unit,
	}

	envelope := events.Envelope{
		Origin:      &originType,
		EventType:   &eventType,
		Deployment:  &deployment,
		Job:         &job,
		Index:       &index,
		Ip:          &ip,
		ValueMetric: &valueMetric,
	}

	server.CacheEnvelope(&envelope)
}
