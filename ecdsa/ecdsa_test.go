package ecdsa

import (
	"encoding/base64"
	"testing"

	"github.com/btcsuite/btcd/btcec"
)

func TestSignAndVerify(t *testing.T) {

	zekePrivKey, zekeKeyPub := btcec.PrivKeyFromBytes(btcec.S256(), zekesPrivKey)
	pubKey1 := base64.URLEncoding.EncodeToString(zekeKeyPub.SerializeCompressed())

	msg := "2GKsZzmRrWTCiwmS29cTfQ==" // base64 encoded string
	sig := Sign(msg, zekePrivKey)

	pubKey, valid, err := VerifyAndExtract(msg, sig, pubKey1)

	if err != nil {
		t.Fatalf(err.Error())
	}
	// sig is valid, and extracted pub key matches
	if valid == false || pubKey1 != pubKey {
		t.Fatalf("nope")
	}
}

var zekesPrivKey = []byte{
	0x2b, 0xd8, 0x07, 0xc9, 0x7f, 0x0e, 0x00, 0xaf,
	0x1a, 0x1f, 0xc3, 0x32, 0x8f, 0xa7, 0x63, 0xa9,
	0x26, 0x97, 0x23, 0xc8, 0xdb, 0x8f, 0xac, 0x4f,
	0x93, 0xaf, 0x72, 0xdb, 0x18, 0x6d, 0x6e, 0x90,
}
