package lsat

import (
	"testing"
)

func TestIsExpiredSatisfier(t *testing.T) {
	now := int64(1663642934209)
	var tests = []struct {
		name string
		expirationValues []string
		creationTime int64
		successFinal bool
		successPrev bool
	}{
		{
			"should pass if time passed is less than expiration",
			[]string{"1000"}, // 1000 seconds from creation
			now - 1000 + 5,
			true,
			true, // not used
		},
		{
			"should fail if time passed is greater than expiration",
			[]string{"1000"}, // 1000 seconds from creation
			now - 10000,
			false,
			true, // not used
		},
		{
			"should pass if successive caveats are increasingly restrictive and last isn't expired",
			[]string{"1000",  "500"}, // last caveat is 500 seconds from creation
			now - 400,
			true, // final value 500 allows current time to pass
			true, 
		},
		{
			"should fail if latter caveat is less restrictive then previous",
			[]string{"500", "1000"}, // 1000 seconds of expiration is more permissive than prev 500, so should fail
			now - 400, 
			true,  // the final value is still valid on its own so this should pass
			false, 
		},
	}

	prefix := MaxUploadCapability
	condition := prefix + CondValidForConstraintSuffix

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			satisfier := IsExpiredSatisfier(prefix, tt.creationTime, now)

			final := tt.expirationValues[len(tt.expirationValues) - 1]
			finalCaveat := NewCaveat(condition, final)
			err := satisfier.SatisfyFinal(finalCaveat)

			if err != nil && tt.successFinal {
				t.Fatal("unexpectedly failed on SatisfyFinal")
			}

			if err == nil && !tt.successFinal {
				t.Fatalf("unexpectedly passed on SatisfyFinal")
			}

			// don't always have a previous caveat, so need to setup vars
			if len(tt.expirationValues) > 1 {
				prev := tt.expirationValues[len(tt.expirationValues) - 2]
				prevCaveat := NewCaveat(condition, prev)

				err := satisfier.SatisfyPrevious(prevCaveat, finalCaveat)
				
				if err != nil && tt.successPrev {
					t.Fatalf("unexpectedly failed on SatisfyPrevious comparing prev %s to current %s", prev, final)
				}

				if err == nil && !tt.successPrev {
					t.Fatalf("unexpectedly passed on SatisfyPrevious comparing prev %s to current %s", prev, final)
				}

			}
		})
	}

}