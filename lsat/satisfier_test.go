package lsat

import (
	"testing"
)

func TestIsExpiredSatisfier(t *testing.T) {
	now := int64(1663642934209)
	var tests = []struct {
		name string
		expirationValues []string
		successFinal bool
		successPrev bool
	}{
		{
			"should pass if current time is before expiration",
			[]string{"1663642935209"}, // 1000 seconds from "current time"
			true,
			true, // not used
		},
		{
			"should fail if time passed is greater than expiration",
			[]string{"1663642933209"}, // 1000 seconds before "current time"
			false,
			true, // not used
		},
		{
			"should pass if successive caveats are increasingly restrictive and last isn't expired",
			// last caveat is 500 seconds after "current time", first caveat is 1000 seconds after
			[]string{"1663642935209",  "1663642934709"}, 
			true, // final value 500 allows current time to pass
			true, 
		},
		{
			"should fail if latter caveat is less restrictive then previous",
			// first caveat is 500 seconds after "current time", last caveat is 1000 seconds after			
			// this is getting less restrictive and should fail
			[]string{"1663642934709", "1663642935209"}, 
			true,  // the final value is still valid on its own so this should pass
			false, 
		},
	}

	prefix := MaxUploadCapability
	condition := prefix + TimeoutConstraintSuffix

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			satisfier := IsExpiredSatisfier(now)

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
