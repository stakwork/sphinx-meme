package main

import (
	"fmt"
	"net/url"

	"github.com/mitchellh/mapstructure"
)

type uploadParams struct {
	Price       int64
	TTL         int64
	Name        string
	Description string
	Tags        []string
	Expiry      int64
}

func decodeForm(vals url.Values, p interface{}) interface{} {
	fields := map[string]interface{}{}
	for k, v := range vals {
		fields[k] = v[0]
	}
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &p,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		fmt.Println(err)
	}
	err = decoder.Decode(fields)

	return &p
}
