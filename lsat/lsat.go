package lsat

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"gopkg.in/macaroon.v2"
)

type contextKey string

var LsatKey = contextKey("lsat")

var (
	authRegex  = regexp.MustCompile("LSAT (.*?):([a-f0-9]{64})")
	authFormat = "LSAT %s:%s"
)

func getLsatAuthorizationHeader(vals []string) (string, error) {
	for _, val := range vals {
		if strings.Contains(val, "LSAT") {
			return val, nil
		}
	}

	return "", errors.New("could not find LSAT in authorization header")
}


// validates auth header and returns macaroon 
// code from aperture library
// https://github.com/lightninglabs/aperture/blob/f9927f3cbe936eebbde4c6c33691116bca850517/lsat/header.go#L54-L62
func validateAuthHeader(authHeader string) (mac string, preimage string, err error) {
		// LSAT tokens come in the format `LSAT [macaroon]:[(preimage)]`
		if !authRegex.MatchString(authHeader) {
			return "", "", fmt.Errorf("invalid "+
				"auth header format: %s", authHeader)
		}

		matches := authRegex.FindStringSubmatch(authHeader)

		if len(matches) != 3 {
			return "", "", fmt.Errorf("invalid "+
				"auth header format: %s", authHeader)
		}

		return matches[1], matches[2], nil
}

func parseLsatHeader(authHeader string) (*macaroon.Macaroon, error) {
	macBase64, _, err := validateAuthHeader(authHeader)

	if err != nil {
		return nil, err
	}

	macBytes, err := base64.StdEncoding.DecodeString(macBase64)

	if err != nil  {
		return nil, fmt.Errorf("base 64 decode of macaroon failed: %v", err)
	}

	mac := &macaroon.Macaroon{}
	err = mac.UnmarshalBinary(macBytes)

	if err != nil {
		return nil, fmt.Errorf("unable to to unmarshal macaroon: %v", err)
	}
	HasCaveat(mac, "foo")
	return mac, nil
}

// simple middleware to get and add LSAT auth header to context
func LsatContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header["Authorization"]
		lsat, err := getLsatAuthorizationHeader(header)
		
		if err != nil {
			fmt.Println("No LSAT authorization header found on request")
			next.ServeHTTP(w, r)
			return
		}
		
		mac, err := parseLsatHeader(lsat)
		
		if err != nil {
			fmt.Printf("could not parse macaroon from authorization header %s", err)
			http.Error(w, http.StatusText(400), 400)
			return
		}

		caveats := mac.Caveats()
		ctx := context.WithValue(r.Context(), LsatKey, caveats)
		next.ServeHTTP(w, r.WithContext((ctx)))
	})
}
