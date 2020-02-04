// Copyright (c) 2016 Blue Medora, Inc. All rights reserved.
// This file is subject to the terms and conditions defined in the included file 'LICENSE.txt'.

package webserver

import (
	"strings"
	"testing"
	"time"
)

func TestTokenStringUniqueness(t *testing.T) {
	t.Log("Generating first token string...")
	tokenStringOne := generateTokenString()

	t.Log("Generating second token string...")
	tokenStringTwo := generateTokenString()

	t.Log("Checking that first and second token strings are not equal...")
	if tokenStringOne == tokenStringTwo {
		t.Errorf("Expected unique strings, received string one %s and string two %s", tokenStringOne, tokenStringTwo)
	}
}

func TestToken(t *testing.T) {
	t.Log("Creating token")
	token := NewToken(tokenTimeoutCallback)
	t.Log("Token Created")

	t.Log("Checking if token is initially valid... (expected value: true)")
	if !token.IsValid() {
		t.Fatalf("Expected is token valid of true, but received false")
	}

	t.Log("Checking if token is intially useable... (expected no-error)")
	err := token.UseToken()
	if err != nil {
		t.Fatalf("Expected token useable, but was un-useable")
	}

	t.Log("Waiting 3 minutes to ensure token invalidates...")
	time.Sleep(time.Duration(3) * time.Minute)

	t.Log("Checking if token is invalid after timeout... (expected value: false)")
	if token.IsValid() {
		t.Fatalf("Expected is token valid of false, but received true")
	}

	t.Log("Checking if token is useable after timeout... (expected error)")
	err = token.UseToken()
	if err == nil {
		t.Fatalf("Expected token un-useable, received error")
	}

	t.Log("Checking if InvalidTokenError... (expected string to contain: Invalid Token Error:)")
	if !strings.Contains(err.Error(), "Invalid Token Error:") {
		t.Errorf("Expected error string to contain 'Invalid Token Error:', but received %s", err.Error())
	}
}

func tokenTimeoutCallback(token *Token) {
	//No logic just need to pass in for test
}
