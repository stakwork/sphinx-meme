package ecdsa

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// VerifyAndExtract ...
func VerifyAndExtract(msg64, sig64 string) (string, bool, error) {

	msg, err := base64.URLEncoding.DecodeString(msg64)
	sig, err := base64.URLEncoding.DecodeString(sig64)
	if err != nil || sig == nil || msg == nil {
		return "", false, errors.New("bad")
	}

	msg = append(signedMsgPrefix, msg...)
	digest := chainhash.DoubleHashB(msg)

	// RecoverCompact both recovers the pubkey and validates the signature.
	pubKey, valid, err := btcec.RecoverCompact(btcec.S256(), sig, digest)
	if err != nil {
		fmt.Printf("ERR: %+v\n", err)
		return "", false, err
	}
	pubKeyZ := base64.URLEncoding.EncodeToString(pubKey.SerializeCompressed())

	return pubKeyZ, valid, nil
}

// Sign ...
func Sign(msgb64 string, privKey *btcec.PrivateKey) string {
	msg, err := base64.URLEncoding.DecodeString(msgb64)
	if err != nil || msg == nil {
		//w.WriteHeader(http.StatusBadRequest)
		return ""
	}

	msg = append(signedMsgPrefix, msg...)
	digest := chainhash.DoubleHashB(msg)
	// btcec.S256(), sig, digest

	sigBytes, err := btcec.SignCompact(btcec.S256(), privKey, digest, true)
	if err != nil {
		return ""
	}

	sig := base64.URLEncoding.EncodeToString(sigBytes)
	return sig
}

var (
	// signedMsgPrefix is a special prefix that we'll prepend to any
	// messages we sign/verify. We do this to ensure that we don't
	// accidentally sign a sighash, or other sensitive material. By
	// prepending this fragment, we mind message signing to our particular
	// context.
	signedMsgPrefix = []byte("Lightning Signed Message:")
)
