package lsat

import (
	"fmt"
	"strconv"
)

// this whole file is just copied from lightning labs' aperture
// these are utility functions for working with caveats
// see source for sample satisfiers:
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


// pass in the file size in the request to compare against
// the caveats
func NewUploadSatisfier(fileSize int16) Satisfier {
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
			if fileSize <= int16(caveatValue) {
				return nil
			}

			return fmt.Errorf("not authorized to upload file with size larger than %d MB", caveatValue)
		},
	}
}