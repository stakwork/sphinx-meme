package ldat

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stakwork/sphinx-meme/ecdsa"
)

func TestTerms(t *testing.T) {

	muid := "qFSOa50yWeGSG8oelsMvctLYdejPRD090dsypBSx_xg="
	pk := "A3PKNqMx2P2EfxkJCHFaNJl7Fdw8XVYMoDLPNBL89JTk"
	var exp uint32 = 1581547570
	host := "memes.sphinx.chat"

	ldat, err := Start(host, muid, pk, exp)
	_, err = Parse(ldat)
	if err != nil {
		fmt.Println(err)
		t.Fatalf("fail")
	}
}

func TestToken(t *testing.T) {

	signingPubKey := "A3PKNqMx2P2EfxkJCHFaNJl7Fdw8XVYMoDLPNBL89JTk"

	tok := "bWVtZS5zcGhpbnguY2hhdA==.qFSOa50yWeGSG8oelsMvctLYdejPRD090dsypBSx_xg=.A5TAzqurrYQm2mZ68JTmPXvsNe1OVYDBc-CWvgzDF8B6.vIlrOg==.IKq12gWXuGqqvwxUZQ5JkCs9akdDLWR1nzSwdUHhKx1rdjQgbmIdxgzrxmlIkhdsfnPOI_-pGB1HsJ8cmN56IPw="
	parsed, _ := Parse(tok)

	sig := parsed.Terms.Sig

	terms := base64.URLEncoding.EncodeToString(parsed.Bytes)

	pubKey, valid, err := ecdsa.VerifyAndExtract(terms, sig)
	if !valid || err != nil {
		t.Fatalf("fail")
		return
	}

	if pubKey != signingPubKey {
		t.Fatalf("damn")
	}
}

func TestTokenWithMeta(t *testing.T) {

	signingPubKey := "A3PKNqMx2P2EfxkJCHFaNJl7Fdw8XVYMoDLPNBL89JTk"

	tok := "bG9jYWxob3N0OjUwMDA=.qFSOa50yWeGSG8oelsMvctLYdejPRD090dsypBSx_xg=.A3PKNqMx2P2EfxkJCHFaNJl7Fdw8XVYMoDLPNBL89JTk.YCxo5w==.YW10PTEwMA==.HwPsHDtW12CQDvvP96pTFcpFORxf0IVq89r4duAcAPOlZx9ElSz8THGPaquyWFbpsR6gN-Ojy6HxXx9XCLEjK2U="
	parsed, _ := Parse(tok)

	fmt.Printf("terms: %+v\n", parsed.Terms)

	sig := parsed.Terms.Sig

	terms := base64.URLEncoding.EncodeToString(parsed.Bytes)

	pubKey, valid, err := ecdsa.VerifyAndExtract(terms, sig)
	if !valid || err != nil {
		t.Fatalf("fail")
		return
	}

	if pubKey != signingPubKey {
		t.Fatalf("damn")
	}
}
