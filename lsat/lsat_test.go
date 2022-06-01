package lsat

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/lightningnetwork/lnd/lntypes"
	"gopkg.in/macaroon.v2"
)

func TestGetLsatAuthorizationHeader(t *testing.T) {
	fakeLsat := "LSAT foobarbaz"
	input1 := []string{"foobar", "foobarbaz", fakeLsat}
	output1, err := getLsatAuthorizationHeader(input1)

	if err != nil {
		t.Errorf("Failed, expected to find LSAT header")
	} else if output1 == fakeLsat {
		t.Logf("Success!")
	}

	input2 := []string{"foobar", "foobarbaz"}
	_, err2 := getLsatAuthorizationHeader(input2)
	
	if err2 == nil {
		t.Errorf("Failed, expected to return error if none found")
	}
}

var lsatPrefix = "LSAT"
var macaroonBase64 = "AgEEbHNhdAJCAAAwpHpumws6ufQoDwiTLNcge0QPUIWA0+tVY+tKPYAJ/zSfmEGlIpNm3VzxuzCqLhEp5KGiyPLUM9L+kcB7uzS+AAIPc2VydmljZXM9bWVtZTowAAISbWVtZV9jYXBhYmlsaXRpZXM9AAAGILA1VCEIExukt4nG+XR9tX8WJ2BVMiHG3UNt1uYJ+NRD"
var preimage = "2ca931a1c36b48f54948b898a271a53ed91ff7d0081939a5fa511249e81cba5c"
func TestParseLsatMacaroon(t *testing.T) {
	testCases := []struct {
		input string
		success bool
		message string
	}{
		{fmt.Sprintf("%s %s:%s", lsatPrefix, macaroonBase64, preimage), true, "expected valid parsing"},
		{fmt.Sprintf("%s:%s", macaroonBase64, preimage), false, "should fail if missing prefix"},
		{fmt.Sprintf("FOO %s:%s", macaroonBase64, preimage), false, "should fail if incorrect prefix"},
		{fmt.Sprintf("%s foo$:%s", lsatPrefix, preimage), false, "should fail if incorrect encoding for macaroon"},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("parseLsatMacaroon=%d", i), func(t*testing.T) {
			_, err := parseLsatHeader(tc.input)
			if err != nil && tc.success {
				t.Fatalf("%s: %s", tc.message, err)
			}
		})
	}
}

func TestLsatContext(t *testing.T) {
	r := chi.NewRouter()
	
	r.Use(LsatContext)
	
	// setup test caveat to make sure it gets added via middleware
	mac := &macaroon.Macaroon{}
	macBytes, _  := base64.StdEncoding.DecodeString(macaroonBase64)
	mac.UnmarshalBinary(macBytes)
	condition := "foo"
	value := "bar"
	caveat := NewCaveat(condition, value)
	rawCaveat := EncodeCaveat(caveat)
	AddFirstPartyCaveats(mac, caveat)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		caveatMap := ctx.Value(ContextKey).(map[string]string)

		if caveatMap == nil {
			t.Fatalf("expected to find lsat caveat map on context")
		} 
		
		if val, ok := caveatMap[condition]; !ok || val != value {
			t.Fatalf("expected key '%s' with value '%s'", condition, value)
		}
	
		w.Write([]byte(rawCaveat))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	if _, body := testRequest(t, ts, "GET", "/", mac); body != rawCaveat {
		t.Fatalf(body)
	}
}

// testing utility copied mostly from go-chi
// https://github.com/go-chi/chi/blob/d32a83448b5f43e42bc96487c6b0b3667a92a2e4/middleware/middleware_test.go#L83
func testRequest(t *testing.T, ts *httptest.Server, method, path string, mac *macaroon.Macaroon) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)

	secret, _ := lntypes.MakePreimageFromStr(preimage)	
	SetHeader(&req.Header, mac, secret)

	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer resp.Body.Close()

	return resp, string(respBody)
}
