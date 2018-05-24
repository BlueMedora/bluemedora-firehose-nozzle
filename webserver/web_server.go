// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package webserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/nozzleconfiguration"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/ttlcache"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/webtoken"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/sonde-go/events"
)

//Webserver Constants
const (
	DefaultCertLocation = "./certs/cert.pem"
	DefaultKeyLocation  = "./certs/key.pem"
	headerUsernameKey   = "username"
	headerPasswordKey   = "password"
	headerTokenKey      = "token"
)

//WebServer REST endpoint for sending data
type WebServer struct {
	logger *gosteno.Logger
	mutext sync.Mutex
	config *nozzleconfiguration.NozzleConfiguration
	tokens map[string]*webtoken.Token //Maps token string to token object
}

//New creates a new WebServer
func New(config *nozzleconfiguration.NozzleConfiguration, logger *gosteno.Logger) *WebServer {
	webserver := WebServer{
		logger: logger,
		config: config,
		tokens: make(map[string]*webtoken.Token),
	}

	webserver.logger.Info("Registering handlers")
	//setup http handlers
	http.HandleFunc("/token", webserver.tokenHandler)
	http.HandleFunc("/metron_agents", webserver.metronAgentsHandler)
	http.HandleFunc("/syslog_drains", webserver.syslogDrainBindersHandler)
	http.HandleFunc("/tps_watchers", webserver.tpsWatcherHandler)
	http.HandleFunc("/tps_listeners", webserver.tpsListenersHandler)
	http.HandleFunc("/stagers", webserver.stagerHandler)
	http.HandleFunc("/ssh_proxies", webserver.sshProxyHandler)
	http.HandleFunc("/senders", webserver.senderHandler)
	http.HandleFunc("/route_emitters", webserver.routeEmitterHandler)
	http.HandleFunc("/reps", webserver.repHandler)
	http.HandleFunc("/receptors", webserver.receptorHandler)
	http.HandleFunc("/nsync_listeners", webserver.nsyncListenerHandler)
	http.HandleFunc("/nsync_bulkers", webserver.nsyncBulkerHandler)
	http.HandleFunc("/garden_linuxs", webserver.gardenLinuxHandler)
	http.HandleFunc("/file_servers", webserver.fileServersHandler)
	http.HandleFunc("/fetchers", webserver.fetcherHandler)
	http.HandleFunc("/convergers", webserver.convergerHandler)
	http.HandleFunc("/cc_uploaders", webserver.ccUploaderHandler)
	http.HandleFunc("/bbs", webserver.bbsHandler)
	http.HandleFunc("/auctioneers", webserver.auctioneerHandler)
	http.HandleFunc("/etcds", webserver.etcdsHandler)
	http.HandleFunc("/doppler_servers", webserver.dopplerServersHandler)
	http.HandleFunc("/cloud_controllers", webserver.cloudControllersHandler)
	http.HandleFunc("/traffic_controllers", webserver.trafficControllersHandler)
	http.HandleFunc("/gorouters", webserver.gorouterHandler)
	http.HandleFunc("/lockets", webserver.locketsHandler)

	return &webserver
}

//Start starts webserver listening
func (webserver *WebServer) Start(keyLocation, certLocation string) <-chan error {
	webserver.logger.Infof("Start listening on port %v", webserver.config.WebServerPort)
	errors := make(chan error, 1)
	go func() {
		defer close(errors)
		if webserver.config.WebServerUseSSL {
			errors <- http.ListenAndServeTLS(fmt.Sprintf(":%v", webserver.config.WebServerPort), getAbsolutePath(certLocation, webserver.logger), getAbsolutePath(keyLocation, webserver.logger), nil)
		} else {
			errors <- http.ListenAndServe(fmt.Sprintf(":%v", webserver.config.WebServerPort), nil)
		}
	}()
	return errors
}

//TokenTimeout is a callback for when a token timesout to remove
func (webserver *WebServer) TokenTimeout(token *webtoken.Token) {
	webserver.mutext.Lock()
	webserver.logger.Debugf("Removing token %s", token.TokenValue)
	delete(webserver.tokens, token.TokenValue)
	webserver.mutext.Unlock()
}

/**Handlers**/
func (webserver *WebServer) tokenHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /token request")
	if r.Method == "GET" {
		username := r.Header.Get(headerUsernameKey)
		password := r.Header.Get(headerPasswordKey)

		//Check for username and password
		if username == "" || password == "" {
			webserver.logger.Debug("No username or password in header")
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "username and/or password not found in header")
		} else {
			//Check validity of username and password
			if username != webserver.config.UAAUsername || password != webserver.config.UAAPassword {
				webserver.logger.Debugf("Wrong username and password for user %s", username)
				w.WriteHeader(http.StatusUnauthorized)
				io.WriteString(w, "Invalid Username and/or Password")
			} else {
				//Successful login
				token := webtoken.New(webserver.TokenTimeout)

				webserver.mutext.Lock()
				webserver.tokens[token.TokenValue] = token
				webserver.mutext.Unlock()

				w.Header().Set(headerTokenKey, token.TokenValue)
				w.WriteHeader(http.StatusOK)

				webserver.logger.Debugf("Successful login generated token <%s>", token.TokenValue)
			}
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, fmt.Sprintf("/token does not support %s http methods", r.Method))
	}
}

func (webserver *WebServer) metronAgentsHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /metron_agents request")
	webserver.processResourceRequest(metronAgentOrigin, w, r)

}

func (webserver *WebServer) syslogDrainBindersHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /syslog_drains request")
	webserver.processResourceRequest(syslogDrainBinderOrigin, w, r)
}

func (webserver *WebServer) tpsWatcherHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /tps_watchers request")
	webserver.processResourceRequest(tpsWatcherOrigin, w, r)
}

func (webserver *WebServer) tpsListenersHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /tps_listeners request")
	webserver.processResourceRequest(tpsListenerOrigin, w, r)
}

func (webserver *WebServer) stagerHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /stagers request")
	webserver.processResourceRequest(stagerOrigin, w, r)
}

func (webserver *WebServer) sshProxyHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /ssh_proxies request")
	webserver.processResourceRequest(sshProxyOrigin, w, r)
}

func (webserver *WebServer) senderHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /senders request")
	webserver.processResourceRequest(senderOrigin, w, r)
}

func (webserver *WebServer) routeEmitterHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /route_emitters request")
	webserver.processResourceRequest(routeEmitterOrigin, w, r)
}

func (webserver *WebServer) repHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /reps request")
	webserver.processResourceRequest(repOrigin, w, r)
}

func (webserver *WebServer) receptorHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /receptors request")
	webserver.processResourceRequest(receptorOrigin, w, r)
}

func (webserver *WebServer) nsyncListenerHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /nsync_listeners request")
	webserver.processResourceRequest(nsyncListenerOrigin, w, r)
}

func (webserver *WebServer) nsyncBulkerHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /nsync_bulkers request")
	webserver.processResourceRequest(nsyncBulkerOrigin, w, r)
}

func (webserver *WebServer) gardenLinuxHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /garden_linuxs request")
	webserver.processResourceRequest(gardenLinuxOrigin, w, r)
}

func (webserver *WebServer) fileServersHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /file_servers request")
	webserver.processResourceRequest(fileServerOrigin, w, r)
}

func (webserver *WebServer) fetcherHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /fetchers request")
	webserver.processResourceRequest(fetcherOrigin, w, r)
}

func (webserver *WebServer) convergerHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /convergers request")
	webserver.processResourceRequest(convergerOrigin, w, r)
}

func (webserver *WebServer) ccUploaderHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /cc_uploaders request")
	webserver.processResourceRequest(ccUploaderOrigin, w, r)
}

func (webserver *WebServer) bbsHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /bbs request")
	webserver.processResourceRequest(bbsOrigin, w, r)
}

func (webserver *WebServer) auctioneerHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /auctioneers request")
	webserver.processResourceRequest(auctioneerOrigin, w, r)
}

func (webserver *WebServer) etcdsHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /etcds request")
	webserver.processResourceRequest(etcdOrigin, w, r)
}

func (webserver *WebServer) dopplerServersHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /doppler_servers request")
	webserver.processResourceRequest(dopplerServerOrigin, w, r)
}

func (webserver *WebServer) cloudControllersHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /cloud_controllers request")
	webserver.processResourceRequest(cloudControllerOrigin, w, r)
}

func (webserver *WebServer) trafficControllersHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /traffic_controllers request")
	webserver.processResourceRequest(trafficControllerOrigin, w, r)
}

func (webserver *WebServer) gorouterHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /gorouters request")
	webserver.processResourceRequest(goRouterOrigin, w, r)
}

func (webserver *WebServer) locketsHandler(w http.ResponseWriter, r *http.Request) {
	webserver.logger.Info("Received /lockets request")
	webserver.processResourceRequest(locketOrigin, w, r)
}

/**Cache Logic**/

//CacheEnvelope caches envelope by origin
func (webserver *WebServer) CacheEnvelope(envelope *events.Envelope) {
	webserver.mutext.Lock()
	defer webserver.mutext.Unlock()

	cache := ttlcache.GetInstance()

	key := createEnvelopeKey(envelope)
	webserver.logger.Debugf("Caching envelope origin %s with key %s", envelope.GetOrigin(), key)

	var resource *ttlcache.Resource
	if value, ok := cache.GetResource(envelope.GetOrigin(), key); ok {
		resource = value
	} else {
		resource = ttlcache.CreateResource(envelope.GetDeployment(), envelope.GetJob(), envelope.GetIndex(), envelope.GetIp())
		cache.SetResource(envelope.GetOrigin(), key, resource)
	}

	resource.AddMetric(envelope, webserver.logger)
}

func (webserver *WebServer) processResourceRequest(originType string, w http.ResponseWriter, r *http.Request) {
	webserver.mutext.Lock()
	defer webserver.mutext.Unlock()

	if r.Method == "GET" {
		tokenString := r.Header.Get(headerTokenKey)

		token := webserver.tokens[tokenString]

		if token == nil || !token.IsTokenValid() {
			webserver.logger.Debugf("Invalid token %s supplied", tokenString)
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, fmt.Sprintf("Invalid token %s supplied", tokenString))
		} else {
			webserver.logger.Debugf("Valid token %s supplied", tokenString)
			token.UseToken()
			webserver.sendOriginBytes(originType, w)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, fmt.Sprintf("Unsupported http method %s", r.Method))
	}
}

func (webserver *WebServer) sendOriginBytes(originType string, w http.ResponseWriter) {
	var messageBytes []byte
	if resourceMap, ok := ttlcache.GetInstance().GetOrigin(originType); ok {
		w.WriteHeader(http.StatusOK)
		values := getValues(resourceMap)
		messageBytes, _ = json.Marshal(values)
	} else {
		w.WriteHeader(http.StatusNoContent)
		messageBytes = []byte("{}")

	}

	_, err := w.Write(messageBytes)

	if err != nil {
		webserver.logger.Errorf("Error while answering end point call for origin %s: %s", originType, err.Error())
	}
}
