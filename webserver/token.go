// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package webserver

import (
	"math/rand"
	"sync"
	"time"
)

const (
	tokenTimeout       = 60
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
	Value                    string
	validToken               bool
	usedSinceLastTimout bool
	timoutTicker             *time.Ticker
	mux                      sync.Mutex
}

//New creates a new token
func NewToken(callback TokenTimeout) *Token {
	newToken := Token{
		Value:               generateTokenString(),
		validToken:               true,
		usedSinceLastTimout: false,
		timoutTicker:             time.NewTicker(time.Duration(tokenTimeout) * time.Second),
	}

	go newToken.startTimeout(callback)

	return &newToken
}

func (t *Token) startTimeout(callback TokenTimeout) {
	for {
		select {
		case <-t.timoutTicker.C:
			t.mux.Lock()
			if !t.usedSinceLastTimout {
				t.validToken = false
				t.mux.Unlock()
				defer callback(t)
				return
			}

			t.usedSinceLastTimout = false
			t.mux.Unlock()
		}
	}
}

func (t *Token) IsValid() bool {
	t.mux.Lock()
	defer t.mux.Unlock()
	return t.validToken
}

func (t *Token) UseToken() error {
	t.mux.Lock()
	defer t.mux.Unlock()

	if t.validToken {
		t.usedSinceLastTimout = true
	} else {
		return &InvalidTokenError{"Attempt to use invalid token"}
	}

	return nil
}

func generateTokenString() string {
	
	tokenLength := 15
	tokenRunes  := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
    
	tokenString := make([]rune, tokenLength)
	for i := range tokenString {
		tokenString[i] = tokenRunes[rand.Intn(len(tokenRunes))]
	}

	return string(tokenString)
}
