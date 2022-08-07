package lsat

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func GetMaxUploadSizeContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.String()

		large := strings.Contains(path, "largefile")
		ctx := r.Context()

		// set default to 32MB in bytes
		maxBytes := int64(32<<20+512)
		
		// Can optional turn on via environment variable the ability to
		// enforce file size restrictions by LSAT. If Aperture is running
		// as a proxy already then the file path will already require an LSAT
		// to access. If aperture is not running but the environment variable
		// is set to true then this will still enforce the size restrictions but only
		// for paths used for large uploads (see routes that call uploadLargeEncryptedFile)
		RESTRICT_UPLOAD_SIZE := os.Getenv("RESTRICT_UPLOAD_SIZE")
		restrictSize, _ := strconv.ParseBool(RESTRICT_UPLOAD_SIZE)
		if RESTRICT_UPLOAD_SIZE != "" && restrictSize  && large {
			size, ok := ctx.Value(MaxUploadSizeContextKey).(int64)
			if !ok {
				fmt.Println("Ran into an error parsing caveat context key")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// convert to bytes w/ buffer
			maxBytes = size << 20 + 512
		} else if RESTRICT_UPLOAD_SIZE != "" && restrictSize {
			// We still want to allow for clients w/o LSAT capability to have some limited upload permissions. 
			// So if they got here not through uploadLargeEncryptedFile but the env variables
			// to enforce restrictions have still been activated then we still want to 
			// cap how big free uploads can be otherwise it's pretty easy to get around the paywall!
			MAX_FREE_UPLOAD_SIZE_MB := os.Getenv("MAX_FREE_UPLOAD_SIZE_MB")
			if MAX_FREE_UPLOAD_SIZE_MB == "" {
				// default to 1MB if none is set as an environment variable
				maxBytes = 1 << 20 + 512
			} else {
				maxFreeMb, err := strconv.ParseInt(MAX_FREE_UPLOAD_SIZE_MB, 10, 16)
				if err != nil {
					fmt.Println("Could not convert free upload size env variable", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				maxBytes = maxFreeMb << 20 + 512
			}
		}

		ctx = context.WithValue(ctx, MaxUploadSizeContextKey, maxBytes)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
