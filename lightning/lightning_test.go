package lightning

import (
	"fmt"
	"testing"
)

func TestConnect(t *testing.T) {
	fmt.Println("hi")
	Init()

	r, err := GetInfo()
	if err != nil {
		t.Fatalf("fail")
		return
	}

	fmt.Printf("%+v\n", r)

	err = SubscribeInvoices()
	if err != nil {
		t.Fatalf("fail")
		return
	}

	for {
	}
}
