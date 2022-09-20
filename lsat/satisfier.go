package lsat

import (
	"fmt"
	"strconv"
)

// This file is based on lightning labs' aperture's satisfiers.
// These are utility functions for working with caveats.
// See source for sample satisfiers:
// https://github.com/lightninglabs/aperture/blob/master/lsat/satisfier.go

// Satisfier provides a generic interface to satisfy a caveat based on its
// condition.
type Satisfier struct {
	// Condition is the condition of the caveat we'll attempt to satisfy.
	Condition string

	// SatisfyPrevious ensures a caveat is in accordance with a previous one
	// with the same condition. This is needed since caveats of the same
	// condition can be used multiple times as long as they enforce more
	// permissions than the previous.
	//
	// For example, we have a caveat that only allows us to use an LSAT for
	// 7 more days. We can add another caveat that only allows for 3 more
	// days of use and lend it to another party.
	SatisfyPrevious func(previous Caveat, current Caveat) error

	// SatisfyFinal satisfies the final caveat of an LSAT. If multiple
	// caveats with the same condition exist, this will only be executed
	// once all previous caveats are also satisfied.
	SatisfyFinal func(Caveat) error
}


// A satisfier that takes in the file size in the request to compare 
// against the caveats and makes sure that the size is less than
// the final caveat and that each matching caveat is increasingly
// restrictive.
func NewUploadSizeSatisfier(fileSize int64) Satisfier {
	return Satisfier {
		// example = large_upload_max_mb
		Condition: MaxUploadCapability + CondMaxUploadConstraintSuffix,

		// confirm that caveat is of increasing restrictiveness 
		SatisfyPrevious: func(prev, cur Caveat) error {
			prevValue, prevValErr := strconv.ParseInt(prev.Value, 10, 16)
			currValue, currValErr := strconv.ParseInt(cur.Value, 10, 16)

			if prevValErr != nil || currValErr != nil {
				return fmt.Errorf("caveat value not a valid integer")
			}

			if currValue > prevValue {
				return fmt.Errorf("%s caveat violates increasing " +
				"restrictiveness", MaxUploadCapability + CondMaxUploadConstraintSuffix)
			}
			
			return nil
		},

		SatisfyFinal: func(c Caveat) error {
			
			caveatValue, err := strconv.ParseInt(c.Value,  10, 16)
			if err != nil {
				// should never reach here 
				return fmt.Errorf("caveat value not a valid integer")
			}
			
			sizeInBytes := caveatValue * (1<<20)
			
			if fileSize <= sizeInBytes {
				return nil
			}

			return fmt.Errorf("not authorized to upload file with size larger than %d MB", caveatValue)
		},
	}
}

// A satisfier to check if an LSAT is expired or not.
// It takes a capability prefix which allows support for different expirations
// for different services or constraints. E.g. you can upload for 6 months
// but can read data for 12 months. 
// The Satisfier takes a creationTime for when the LSAT was created to compare against the expiration(s) 
// in the caveats and the currentTimestamp which is passed in as the third argument. The satisfier
// will also make sure that each subsequent caveat of the same condition only has increasingly
// strict expirations.
func IsExpiredSatisfier(capabilityPrefix string, creationTime int64, currentTimestamp int64) Satisfier {
	return Satisfier {
		Condition: capabilityPrefix + CondValidForConstraintSuffix,

		// check that each caveat is more restrictive than the previous
		SatisfyPrevious: func(prev, cur Caveat) error {
			prevValue, prevValErr := strconv.ParseInt(prev.Value, 10, 16)
			currValue, currValErr := strconv.ParseInt(cur.Value, 10, 16)
			if prevValErr != nil || currValErr != nil {
				return fmt.Errorf("caveat value not a valid integer")
			}
			
			// to be valid, each subsequent expiration time must be sooner (smaller) than previous
			if currValue > prevValue {
				return fmt.Errorf("%s caveat violates increasing " +
				"restrictiveness", capabilityPrefix + CondValidForConstraintSuffix)
			}
			
			return nil
		},

		// make sure that the final relevant caveat is not passed
		// the current date/time
		SatisfyFinal: func(c Caveat) error {

			secondsUntilExpiration, err := strconv.ParseInt(c.Value, 10, 16)

			if err != nil {
				// should never reach here 
				return fmt.Errorf("caveat value not a valid integer")
			}

			if currentTimestamp - creationTime < secondsUntilExpiration {
				return nil
			}
			
			return fmt.Errorf("not authorized to upload file. LSAT has expired")
		},
	}
}
