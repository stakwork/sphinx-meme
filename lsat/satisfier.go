package lsat

import (
	"fmt"
	"strconv"
	"time"
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
// The Satisfier takes a expirationTime for when the LSAT was created to compare against the expiration(s) 
// in the caveats and the currentTimestamp which is passed in as the third argument. The satisfier
// will also make sure that each subsequent caveat of the same condition only has increasingly
// strict expirations.
func IsExpiredSatisfier(currentTimestamp int64) Satisfier {
	return Satisfier {
		Condition: MemeServerServicePrefix + TimeoutConstraintSuffix,

		// check that each caveat is more restrictive than the previous
		SatisfyPrevious: func(prev, cur Caveat) error {
			prevValue, prevValErr := strconv.ParseInt(prev.Value, 10, 64)
			currValue, currValErr := strconv.ParseInt(cur.Value, 10, 64)
			
			if prevValErr != nil || currValErr != nil {
				return fmt.Errorf("caveat value not a valid integer")
			}
			
			prevTime := time.Unix(prevValue, 0)
			currTime := time.Unix(currValue, 0)
			
			// Satisfier should fail if a previous timestamp in the list is
			// earlier than ones after it b/c that means they are getting
			// more permissive. 
			if prevTime.Before(currTime) {
				return fmt.Errorf("%s caveat violates increasing " +
				"restrictiveness", MemeServerServicePrefix + TimeoutConstraintSuffix)
			}
			
			return nil
		},

		// make sure that the final relevant caveat is not passed
		// the current date/time
		SatisfyFinal: func(c Caveat) error {

			expirationTimestamp, err := strconv.ParseInt(c.Value, 10, 64)
			
			if err != nil {
				// should never reach here 
				return fmt.Errorf("caveat value not a valid integer")
			}

			expirationTime := time.Unix(expirationTimestamp, 0)
			currentTime := time.Unix(currentTimestamp, 0)
			
			if currentTime.Before(expirationTime) {
				return nil
			}
			
			return fmt.Errorf("not authorized to upload file. LSAT has expired")
		},
	}
}
