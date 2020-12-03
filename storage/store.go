package storage

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Init starts
func Init() {
	mode := os.Getenv("STORAGE_MODE")
	if mode == "" {
		mode = "local"
	}
	if mode == "s3" || mode == "S3" {
		Store = &space
	} else {
		Store = &Local // pointer
	}
	fmt.Printf("storage mode: %s\n", mode)
	Store.Init()
}

// Store ...
var Store store

type store interface {
	Init()
	GetReader(string, [32]byte) (io.ReadCloser, error)
	PostReader(string, *bytes.Buffer, int64, string, [32]byte) error
	Delete(string) error
	GenNonce() ([32]byte, error)
	List(string) ([]string, error)
}
