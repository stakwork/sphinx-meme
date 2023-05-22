package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/stakwork/sphinx-meme/auth"
	"github.com/stakwork/sphinx-meme/ecdsa"
)

// TIMEOUT is the number of seconds until req becomes invalid
var TIMEOUT = 10

func ask(w http.ResponseWriter, r *http.Request) {
	ts := strconv.Itoa(int(time.Now().Unix()))
	// h := blake2b.Sum256([]byte(ts))
	h := []byte(ts)
	challenge := base64.URLEncoding.EncodeToString(h[:])

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"id":        ts,
		"challenge": challenge,
	})
}

func verify(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.FormValue("id")
	sig := r.FormValue("sig")
	pubkey := r.FormValue("pubkey")
	readonly := r.FormValue("readonly")
	fmt.Printf("id %s\n", id)
	fmt.Printf("sig %s\n", sig)
	fmt.Printf("pubkey %s\n", pubkey)
	if id == "" || sig == "" {
		fmt.Println("=> no sig or id")
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	ts, err := strconv.Atoi(id)
	if err != nil || ts == 0 {
		fmt.Println("invalid ts")
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	now := int(time.Now().Unix())
	// deny requests that are too old or from the future
	if ts <= now-TIMEOUT || ts > now {
		fmt.Println("too late")
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	// h := blake2b.Sum256([]byte(id))
	h := []byte(id)
	challenge := base64.URLEncoding.EncodeToString(h[:])

	pkb, _ := hex.DecodeString(pubkey)
	expectedPubky := base64.URLEncoding.EncodeToString(pkb)

	pubKeyExtracted, valid, err := ecdsa.VerifyAndExtract(challenge, sig, expectedPubky)
	if !valid || err != nil {
		fmt.Println("not verified")
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	claims := jwt.MapClaims{
		"key": pubKeyExtracted,
		"exp": auth.ExpireInHours(24 * 7),
	}
	if readonly != "" {
		claims["readonly"] = true
	}
	fmt.Printf("CLAIMS: %+v\n", claims)
	_, tokenString, err := auth.TokenAuth.Encode(claims)
	if err != nil {
		fmt.Println("error creating JWT")
		w.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}
