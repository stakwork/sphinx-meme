package lsat

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime/multipart"
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
	condition := "foo"
	value := "bar"
	mac, rawCaveat := getMacaroon(condition, value)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		caveats := ctx.Value(ContextKey).([]Caveat)
		if caveats == nil {
			t.Fatalf("expected to find lsat caveat map on context")
		} 
		
		found := false
		for _, c := range caveats {
			if c.Condition == condition {
				found = true
			}
		}

		if !found {
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

func TestVerifyUploadContext(t *testing.T) {
	var tests = []struct {
		name string
		values []string
		fileSize int16
		want int
	}{
		{
			"should accept requests if the file size is below caveat value",
			[]string{"2"},
			1,
			http.StatusOK,
		},
		{
			"should reject requests if file size is larger than caveat value",
			[]string{"1"},
			2,
			http.StatusUnauthorized,
		},
		{
			"should reject requests if caveats are not of increasing restrictiveness",
			[]string{"1", "2"},
			1,
			http.StatusUnauthorized,
		},
	}
	caveatCondition := MaxUploadCapability + CondMaxUploadConstraintSuffix // large_upload_max_mb 
	
	for _, tt := range tests  {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			// setup macaroon with caveats, adding the first caveat on the macaroon
			val, values := tt.values[0], tt.values[1:]
			mac, _ := getMacaroon(caveatCondition, val)
			
			// add other caveats if there are any
			if len(values) > 0 {
				for _, v := range values {
					c := NewCaveat(caveatCondition, v)
					AddFirstPartyCaveats(mac, c)
				}
			}
			
			r := chi.NewRouter()
			// middleware to make sure the caveats are on the context
			r.Use(LsatContext)
			// verify the caveats from the context
			r.Use(VerifyUploadContext)
			// a nothing route that we can call to go through the relevant middleware
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {})

			// mocking file upload in the request
			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", "file.png")
			b := make([]byte, tt.fileSize)
			part.Write(b)
			writer.Close()

			// setup request
			req := httptest.NewRequest("POST", "/", body)
			req.Header.Add("Content-Type", "multipart/form-data")
			req.Header.Set("Content-Type", writer.FormDataContentType())
			secret, _ := lntypes.MakePreimageFromStr(preimage)	
			SetHeader(&req.Header, mac, secret)

			r.ServeHTTP(recorder, req)
			res := recorder.Result()

			if res.StatusCode != tt.want {
				t.Errorf("response is incorrect, got %d, want %d", recorder.Code, tt.want)
			}
		})

	}
}

func getMacaroon(condition string, value string) (mac *macaroon.Macaroon, rawCaveat string) {
	// setup test caveat to make sure it gets added via middleware
	mac = &macaroon.Macaroon{}
	macBytes, _  := base64.StdEncoding.DecodeString(macaroonBase64)
	mac.UnmarshalBinary(macBytes)
	caveat := NewCaveat(condition, value)
	rawCaveat = EncodeCaveat(caveat)
	AddFirstPartyCaveats(mac, caveat)

	return
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
