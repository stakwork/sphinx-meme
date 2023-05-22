package ldat

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func tsBytes() []byte {
	b := make([]byte, 4)
	t := time.Now().UTC().Unix()
	fmt.Println(t)
	binary.BigEndian.PutUint32(b, uint32(t))
	return b
}

// func TestUtf8(t *testing.T) {
// 	b := tsBytes()
// 	if !utf8.Valid(b) {
// 		t.Fatalf("not utf8")
// 	}
// }

func TestSignAndVerifyBothy(t *testing.T) {

	// setup signer
	zekePrivKey, zekeKeyPub := btcec.PrivKeyFromBytes(btcec.S256(), zekesPrivKey)
	pubKey1 := base64.URLEncoding.EncodeToString(zekeKeyPub.SerializeCompressed())

	// start ldat
	muid := "qFSOa50yWeGSG8oelsMvctLYdejPRD090dsypBSx_xg="
	pk := "A3PKNqMx2P2EfxkJCHFaNJl7Fdw8XVYMoDLPNBL89JTk"
	var exp uint32 = 1581547570
	host := "memes.sphinx.chat"
	ldatstart, err := Start(host, muid, pk, exp)
	if err != nil {
		t.Fatalf(err.Error())
	}
	ldat, err := Parse(ldatstart)
	if err != nil {
		fmt.Println(err)
		t.Fatalf("fail")
	}

	// sign raw
	sig := Sign(ldat.Bytes, zekePrivKey)
	msg := base64.URLEncoding.EncodeToString(ldat.Bytes)
	pk1, valid, err := VerifyAndExtract(msg, sig, pubKey1)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// raw sig is valid, and extracted pub key matches
	if valid == false {
		t.Fatalf("invalid")
	}
	if pk1 != pubKey1 {
		t.Fatalf("invalid pubkey")
	}

	// sign utf8 bytes of base64
	msg2 := base64.URLEncoding.EncodeToString(ldat.Bytes)
	sig2 := Sign([]byte(msg2), zekePrivKey)
	pk2, valid2, err := VerifyAndExtract(msg2, sig2, pubKey1)
	if err != nil {
		t.Fatalf(err.Error())
	}
	// utf8 sig is valid, and extracted pub key matches
	if valid2 == false {
		t.Fatalf("invalid 2")
	}
	if pk2 != pubKey1 {
		t.Fatalf("invalid pubkey 2")
	}
}

var zekesPrivKey = []byte{
	0x2b, 0xd8, 0x07, 0xc9, 0x7f, 0x0e, 0x00, 0xaf,
	0x1a, 0x1f, 0xc3, 0x32, 0x8f, 0xa7, 0x63, 0xa9,
	0x26, 0x97, 0x23, 0xc8, 0xdb, 0x8f, 0xac, 0x4f,
	0x93, 0xaf, 0x72, 0xdb, 0x18, 0x6d, 0x6e, 0x90,
}

func stringify(a string) ([]byte, error) {
	return []byte(a), nil
}

// VerifyAndExtract ...
func VerifyAndExtract(msg64, sig64, expectedPubKey string) (string, bool, error) {

	err := errors.New("invalid sig")
	for _, decoder := range []func(string) ([]byte, error){
		base64.URLEncoding.DecodeString,
		stringify,
	} {
		msg, err := decoder(msg64)
		if err == nil {
			pubKey64, valid, err := verifyAndExtractInner(msg, sig64)
			if err != nil {
				return "", false, err
			}
			if expectedPubKey != "" {
				if pubKey64 != expectedPubKey {
					err = errors.New("unexpected pubkey")
					continue // skip to next one
					// if both "continue" then it fails
				}
			}
			return pubKey64, valid, nil
		}
	}
	return "", false, err
}

func verifyAndExtractInner(msg []byte, sig64 string) (string, bool, error) {
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
	pubKey64 := base64.URLEncoding.EncodeToString(pubKey.SerializeCompressed())
	return pubKey64, valid, nil
}

// Sign ...
func Sign(msg []byte, privKey *btcec.PrivateKey) string {
	if msg == nil {
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
