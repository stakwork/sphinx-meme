package lsat

import (
	"fmt"
	"testing"
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
		// {fmt.Sprintf("%s %s:", lsatPrefix, macaroonBase64), true, "expected valid parsing"},
		{fmt.Sprintf("%s:%s", macaroonBase64, preimage), false, "should fail if missing prefix"},
		{fmt.Sprintf("FOO %s:%s", macaroonBase64, preimage), false, "should fail if incorrect prefix"},
		{fmt.Sprintf("%s foo$:%s", lsatPrefix, preimage), false, "should fail if incorrect encoding for macaroon"},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("parseLsatmacaroon=%d", i), func(t*testing.T) {
			_, err := parseLsatHeader(tc.input)
			if err != nil && tc.success {
				t.Fatalf("%s: %s", tc.message, err)
			}
		})
	}
}