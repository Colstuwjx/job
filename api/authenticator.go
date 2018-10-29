// Copyright Project Harbor Authors. All rights reserved.

package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Colstuwjx/job/config"
	"github.com/Colstuwjx/job/utils"
)

const (
	authHeader = "Authorization"
)

// Authenticator defined behaviors of doing auth checking.
type Authenticator interface {
	// DoAuth auth incoming request
	//
	// req *http.Request: the incoming request
	//
	// Returns:
	// nil returned if successfully done
	// otherwise an error returned
	DoAuth(req *http.Request) error
}

// SecretAuthenticator implements interface 'Authenticator' based on simple secret.
type SecretAuthenticator struct{}

// DoAuth implements same method in interface 'Authenticator'.
func (sa *SecretAuthenticator) DoAuth(req *http.Request) error {
	if req == nil {
		return errors.New("nil request")
	}

	secret := strings.TrimSpace(req.Header.Get(authHeader))
	if utils.IsEmptyStr(secret) {
		return fmt.Errorf("header '%s' missing", authHeader)
	}

	expectedSecret := config.GetUIAuthSecret()
	if secret != expectedSecret {
		return errors.New("unauthorized")
	}

	return nil
}
