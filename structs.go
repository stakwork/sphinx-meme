package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

// Media struct
type Media struct {
	ID          string         `json:"muid"`
	OwnerPubKey string         `json:"owner_pub_key"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Price       int64          `json:"price"`
	Tags        pq.StringArray `json:"tags"`
	Filename    string         `json:"filename"`
	TTL         int64          `json:"ttl"`
	Size        int64          `json:"size"`
	Mime        string         `json:"mime"`
	Nonce       string         `json:"-"`
	Created     *time.Time     `json:"created"`
	Updated     *time.Time     `json:"updated"`
	Expiry      *time.Time     `json:"expiry"`
	TotalSats   int64          `json:"total_sats,omitempty"`
	TotalBuys   int64          `json:"total_buys,omitempty"`
	Width       int            `json:"width"`
	Height      int            `json:"height"`
	Template    bool           `json:"template"`
}

// PropertyMap ...
type PropertyMap map[string]interface{}

// Value ...
func (p PropertyMap) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// Scan ...
func (p *PropertyMap) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	var i interface{}
	if err := json.Unmarshal(source, &i); err != nil {
		return err
	}

	*p, ok = i.(map[string]interface{})
	if !ok {
		return errors.New("type assertion .(map[string]interface{}) failed")
	}

	return nil
}
