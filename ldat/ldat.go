package ldat

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Terms ... All strings are base64 encoded
type Terms struct {
	Host        string
	Muid        string
	Exp         int64 // unix timestamp of expiry
	BuyerPubKey string
	Sig         string
	Meta        Meta
}

// Start ...
func Start(host, muid, pubkey string, exp uint32) (string, error) {
	hostBuf := []byte(host)
	muidBuf, err := base64.URLEncoding.DecodeString(muid)
	if err != nil {
		return "", err
	}
	pubkeyBuf, err := base64.URLEncoding.DecodeString(pubkey)
	if err != nil {
		return "", err
	}
	expBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(expBuf, exp)

	as := [][]byte{hostBuf, muidBuf, pubkeyBuf, expBuf}
	s := ""
	for _, a := range as {
		s += base64.URLEncoding.EncodeToString(a)
		s += "."
	}
	return s, nil
}

// ParsedTerms ...
type ParsedTerms struct {
	Terms Terms
	Bytes []byte
}

// Parse ...
func Parse(token string) (ParsedTerms, error) {
	fmt.Printf("token %s\n", token)
	ta := strings.Split(token, ".")
	if len(ta) < 5 {
		return ParsedTerms{}, errors.New("too short")
	}
	termz := Terms{}
	ba := []byte{}
	for i, section := range ta {
		b, err := base64.URLEncoding.DecodeString(section)
		if err != nil {
			return ParsedTerms{}, err
		}
		if i != len(ta)-1 { // dont sign the sig of course
			ba = append(ba, b...)
		}
		switch i {
		case 0:
			termz.Host = string(b)
		case 1:
			termz.Muid = base64.URLEncoding.EncodeToString(b)
		case 2:
			termz.BuyerPubKey = base64.URLEncoding.EncodeToString(b)
		case 3:
			termz.Exp = int64(binary.BigEndian.Uint32(b))
		case len(ta) - 1:
			termz.Sig = section
		default:
			termz.Meta = parseMeta(b)
		}
	}
	return ParsedTerms{
		Terms: termz,
		Bytes: ba,
	}, nil
}

// Meta struct
type Meta map[string]string

func parseMeta(b []byte) Meta {
	params, err := url.ParseQuery(string(b))
	if err != nil {
		return Meta{}
	}
	meta := Meta{}
	for k, v := range params {
		if len(v) > 0 {
			meta[k] = v[0]
		}
	}
	return meta
}
