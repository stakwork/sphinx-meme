package auth

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/jwtauth"
)

type contextKey string

// ContextKey for public key
var (
	ContextKey  = contextKey("key")
	ContextHost = contextKey("host")
	defaultHost = "localhost:5005"
)

// Verifier ...
func Verifier(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return jwtauth.Verify(ja, jwtauth.TokenFromQuery, jwtauth.TokenFromHeader, jwtauth.TokenFromCookie)(next)
	}
}

// TokenAuth is a global authenticator interface
var TokenAuth *jwtauth.JWTAuth

// ExpireInHours for jwt
func ExpireInHours(hours int) int64 {
	return jwtauth.ExpireIn(time.Duration(hours) * time.Hour)
}

// Init auth
func Init() {
	jwtKey := os.Getenv("JWT_KEY")
	if jwtKey == "" {
		log.Fatal("No JWT key")
	}
	TokenAuth = jwtauth.New("HS256", []byte(jwtKey), nil)
}

func getHost() string {
	host := os.Getenv("HOST")
	if host == "" {
		host = defaultHost
	}
	return host
}

// PubKeyContext parses JWT fields
func PubKeyContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, _ := jwtauth.FromContext(r.Context())
		pubKey, ok := claims["key"].(string)
		if !ok || pubKey == "" {
			http.Error(w, http.StatusText(401), 401)
			return
		}
		ctx := context.WithValue(r.Context(), ContextKey, pubKey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HostContext ...
func HostContext(next http.Handler) http.Handler {
	host := getHost()
	//fmt.Println(host)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ContextHost, host)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
