// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package webtoken

import (
	"sync"
	"time"
)

var (
	tokenTimeout = 60
)

//TokenTimeout callback when a token times out
type TokenTimeout func(token *Token)

//InvalidTokenError signals invalid token usage
type InvalidTokenError struct {
	s string
}

func (e *InvalidTokenError) Error() string {
	return "Invalid Token Error: " + e.s
}

//Token token used for webserver communication
type Token struct {
	TokenValue               string
	validToken               bool
	tokenUsedSinceLastTimout bool
	timoutTicker             *time.Ticker
	mux                      sync.Mutex
}

//New creates a new token
func New(timeoutCallback TokenTimeout) *Token {
	newToken := Token{
		TokenValue:               GenerateTokenString(),
		validToken:               true,
		tokenUsedSinceLastTimout: false,
		timoutTicker:             time.NewTicker(time.Duration(tokenTimeout) * time.Second),
	}

	go newToken.startTimeout(timeoutCallback)

	return &newToken
}

func (token *Token) startTimeout(timeoutCallback TokenTimeout) {
	for {
		select {
		case <-token.timoutTicker.C:
			token.mux.Lock()
			if !token.tokenUsedSinceLastTimout {
				token.validToken = false
				token.mux.Unlock()
				defer timeoutCallback(token)
				return
			}

			token.tokenUsedSinceLastTimout = false
			token.mux.Unlock()
		}
	}
}

//IsTokenValid validity of token
func (token *Token) IsTokenValid() bool {
	token.mux.Lock()
	defer token.mux.Unlock()
	return token.validToken
}

//UseToken marks token as used
func (token *Token) UseToken() error {
	token.mux.Lock()
	defer token.mux.Unlock()

	if token.validToken {
		token.tokenUsedSinceLastTimout = true
	} else {
		return &InvalidTokenError{"Attempt to use invalid token"}
	}

	return nil
}
