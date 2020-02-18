// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package webserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/configuration"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/results"
	"github.com/BlueMedoraPublic/bluemedora-firehose-nozzle/ttlcache"

	"github.com/cloudfoundry/gosteno"
)

//Webserver Constants
const (
	headerUsernameKey = "username"
	headerPasswordKey = "password"
	headerTokenKey    = "token"
)

//WebServer REST endpoint for sending data
type WebServer struct {
	sync.Mutex
	logger *gosteno.Logger
	config *configuration.Configuration
	tokens map[string]*Token //Maps token string to token object
}

//New creates a new WebServer
func New(c *configuration.Configuration, l *gosteno.Logger) *WebServer {
	ws := &WebServer{
		logger: l,
		config: c,
		tokens: make(map[string]*Token),
	}

	ws.logger.Info("Registering handlers")
	//setup http handlers
	http.HandleFunc("/token", ws.tokenHandler)
	http.HandleFunc("/metron_agents", ws.metronAgentsHandler)
	http.HandleFunc("/syslog_drains", ws.syslogDrainBindersHandler)
	http.HandleFunc("/tps_watchers", ws.tpsWatcherHandler)
	http.HandleFunc("/tps_listeners", ws.tpsListenersHandler)
	http.HandleFunc("/stagers", ws.stagerHandler)
	http.HandleFunc("/ssh_proxies", ws.sshProxyHandler)
	http.HandleFunc("/senders", ws.senderHandler)
	http.HandleFunc("/route_emitters", ws.routeEmitterHandler)
	http.HandleFunc("/reps", ws.repHandler)
	http.HandleFunc("/receptors", ws.receptorHandler)
	http.HandleFunc("/nsync_listeners", ws.nsyncListenerHandler)
	http.HandleFunc("/nsync_bulkers", ws.nsyncBulkerHandler)
	http.HandleFunc("/garden_linuxs", ws.gardenLinuxHandler)
	http.HandleFunc("/file_servers", ws.fileServersHandler)
	http.HandleFunc("/fetchers", ws.fetcherHandler)
	http.HandleFunc("/convergers", ws.convergerHandler)
	http.HandleFunc("/cc_uploaders", ws.ccUploaderHandler)
	http.HandleFunc("/bbs", ws.bbsHandler)
	http.HandleFunc("/auctioneers", ws.auctioneerHandler)
	http.HandleFunc("/etcds", ws.etcdsHandler)
	http.HandleFunc("/doppler_servers", ws.dopplerServersHandler)
	http.HandleFunc("/cloud_controllers", ws.cloudControllersHandler)
	http.HandleFunc("/traffic_controllers", ws.trafficControllersHandler)
	http.HandleFunc("/gorouters", ws.gorouterHandler)
	http.HandleFunc("/lockets", ws.locketsHandler)

	return ws
}

func (ws *WebServer) Start() <-chan error {
	ws.logger.Infof("Start listening on port %v", ws.config.WebServerPort)
	errors := make(chan error, 1)
	go func() {
		defer close(errors)
		if ws.config.WebServerUseSSL {
			errors <- http.ListenAndServeTLS(fmt.Sprintf(":%v", ws.config.WebServerPort), getAbsolutePath(ws.config.WebServerCertLocation, ws.logger), getAbsolutePath(ws.config.WebServerKeyLocation, ws.logger), nil)
		} else {
			errors <- http.ListenAndServe(fmt.Sprintf(":%v", ws.config.WebServerPort), nil)
		}
	}()
	return errors
}

func (ws *WebServer) TokenTimeout(token *Token) {
	ws.Lock()
	defer ws.Unlock()
	ws.logger.Debugf("Removing token %s", token.Value)
	delete(ws.tokens, token.Value)
}

/**Handlers**/
func (ws *WebServer) tokenHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /token request")
	if r.Method == "GET" {
		username := r.Header.Get(headerUsernameKey)
		password := r.Header.Get(headerPasswordKey)

		//Check for username and password
		if username == "" || password == "" {
			ws.logger.Debug("No username or password in header")
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "username and/or password not found in header")
		} else {
			//Check validity of username and password
			if username != ws.config.UAAUsername || password != ws.config.UAAPassword {
				ws.logger.Debugf("Wrong username and password for user %s", username)
				w.WriteHeader(http.StatusUnauthorized)
				io.WriteString(w, "Invalid Username and/or Password")
			} else {
				//Successful login
				token := NewToken(ws.TokenTimeout)

				ws.Lock()
				ws.tokens[token.Value] = token
				ws.Unlock()

				w.Header().Set(headerTokenKey, token.Value)
				w.WriteHeader(http.StatusOK)

				ws.logger.Debugf("Successful login generated token <%s>", token.Value)
			}
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, fmt.Sprintf("/token does not support %s http methods", r.Method))
	}
}

func (ws *WebServer) metronAgentsHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /metron_agents request")
	ws.processResourceRequest(metronAgentOrigin, w, r)
}

func (ws *WebServer) syslogDrainBindersHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /syslog_drains request")
	ws.processResourceRequest(syslogDrainBinderOrigin, w, r)
}

func (ws *WebServer) tpsWatcherHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /tps_watchers request")
	ws.processResourceRequest(tpsWatcherOrigin, w, r)
}

func (ws *WebServer) tpsListenersHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /tps_listeners request")
	ws.processResourceRequest(tpsListenerOrigin, w, r)
}

func (ws *WebServer) stagerHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /stagers request")
	ws.processResourceRequest(stagerOrigin, w, r)
}

func (ws *WebServer) sshProxyHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /ssh_proxies request")
	ws.processResourceRequest(sshProxyOrigin, w, r)
}

func (ws *WebServer) senderHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /senders request")
	ws.processResourceRequest(senderOrigin, w, r)
}

func (ws *WebServer) routeEmitterHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /route_emitters request")
	ws.processResourceRequest(routeEmitterOrigin, w, r)
}

func (ws *WebServer) repHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /reps request")
	ws.processResourceRequest(repOrigin, w, r)
}

func (ws *WebServer) receptorHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /receptors request")
	ws.processResourceRequest(receptorOrigin, w, r)
}

func (ws *WebServer) nsyncListenerHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /nsync_listeners request")
	ws.processResourceRequest(nsyncListenerOrigin, w, r)
}

func (ws *WebServer) nsyncBulkerHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /nsync_bulkers request")
	ws.processResourceRequest(nsyncBulkerOrigin, w, r)
}

func (ws *WebServer) gardenLinuxHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /garden_linuxs request")
	ws.processResourceRequest(gardenLinuxOrigin, w, r)
}

func (ws *WebServer) fileServersHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /file_servers request")
	ws.processResourceRequest(fileServerOrigin, w, r)
}

func (ws *WebServer) fetcherHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /fetchers request")
	ws.processResourceRequest(fetcherOrigin, w, r)
}

func (ws *WebServer) convergerHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /convergers request")
	ws.processResourceRequest(convergerOrigin, w, r)
}

func (ws *WebServer) ccUploaderHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /cc_uploaders request")
	ws.processResourceRequest(ccUploaderOrigin, w, r)
}

func (ws *WebServer) bbsHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /bbs request")
	ws.processResourceRequest(bbsOrigin, w, r)
}

func (ws *WebServer) auctioneerHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /auctioneers request")
	ws.processResourceRequest(auctioneerOrigin, w, r)
}

func (ws *WebServer) etcdsHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /etcds request")
	ws.processResourceRequest(etcdOrigin, w, r)
}

func (ws *WebServer) dopplerServersHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /doppler_servers request")
	ws.processResourceRequest(dopplerServerOrigin, w, r)
}

func (ws *WebServer) cloudControllersHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /cloud_controllers request")
	ws.processResourceRequest(cloudControllerOrigin, w, r)
}

func (ws *WebServer) trafficControllersHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /traffic_controllers request")
	ws.processResourceRequest(trafficControllerOrigin, w, r)
}

func (ws *WebServer) gorouterHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /gorouters request")
	ws.processResourceRequest(goRouterOrigin, w, r)
}

func (ws *WebServer) locketsHandler(w http.ResponseWriter, r *http.Request) {
	ws.logger.Info("Received /lockets request")
	ws.processResourceRequest(locketOrigin, w, r)
}

func (ws *WebServer) processResourceRequest(originType string, w http.ResponseWriter, r *http.Request) {
	ws.Lock()
	defer ws.Unlock()

	if r.Method == "GET" {
		tokenString := r.Header.Get(headerTokenKey)

		token := ws.tokens[tokenString]

		if token != nil && token.IsValid() {
			ws.logger.Debugf("Valid token %s supplied", tokenString)
			token.UseToken()
			ws.sendOriginBytes(originType, w)
		} else {
			ws.logger.Debugf("Invalid token %s supplied", tokenString)
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, fmt.Sprintf("Invalid token %s supplied", tokenString))

		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, fmt.Sprintf("Unsupported http method %s", r.Method))
	}
}

func (ws *WebServer) sendOriginBytes(originType string, w http.ResponseWriter) {
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
		ws.logger.Errorf("Error while answering end point call for origin %s: %s", originType, err.Error())
	}
}

const (
	metronAgentOrigin       = "MetronAgent"
	syslogDrainBinderOrigin = "syslog_drain_binder"
	tpsWatcherOrigin        = "tps_watcher"
	tpsListenerOrigin       = "tps_listener"
	stagerOrigin            = "stager"
	sshProxyOrigin          = "ssh-proxy"
	senderOrigin            = "sender"
	routeEmitterOrigin      = "route_emitter"
	repOrigin               = "rep"
	receptorOrigin          = "receptor"
	nsyncListenerOrigin     = "nsync_listener"
	nsyncBulkerOrigin       = "nsync_bulker"
	gardenLinuxOrigin       = "garden-linux"
	fileServerOrigin        = "file_server"
	fetcherOrigin           = "fetcher"
	convergerOrigin         = "converger"
	ccUploaderOrigin        = "cc_uploader"
	bbsOrigin               = "bbs"
	auctioneerOrigin        = "auctioneer"
	etcdOrigin              = "etcd"
	dopplerServerOrigin     = "DopplerServer"
	cloudControllerOrigin   = "cc"
	trafficControllerOrigin = "LoggregatorTrafficController"
	goRouterOrigin          = "gorouter"
	locketOrigin            = "locket"
)

func getValues(resourceMap map[string]*results.Resource) []*results.Resource {
	resources := make([]*results.Resource, 0, len(resourceMap))

	for _, resource := range resourceMap {
		resources = append(resources, resource)
	}

	return resources
}

func getAbsolutePath(file string, logger *gosteno.Logger) string {
	logger.Infof("Finding absolute path to %s", file)
	absolutePath, err := filepath.Abs(file)

	if err != nil {
		logger.Warnf("Error getting absolute path to $s using relative path due to %v", file)
		return file
	}

	return absolutePath
}
