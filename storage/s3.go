package storage

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/s3"
)

type s3store struct {
	bucket *s3.Bucket
	prefix string
}

var space s3store

func (store s3store) Init() {
	s3Key := os.Getenv("S3_KEY")
	s3Secret := os.Getenv("S3_SECRET")
	if s3Key == "" || s3Secret == "" {
		log.Panic("MISSING S3 CREDS!!!!")
		return
	}
	AWSAuth := aws.Auth{
		AccessKey: s3Key,
		SecretKey: s3Secret,
	}
	// https://github.com/goamz/goamz/blob/master/aws/regions.go
	connection := s3.New(AWSAuth, aws.USEast)
	space.bucket = connection.Bucket("sphinx-memes")
	space.prefix = ""
}

func (store s3store) Delete(path string) error {
	err := store.bucket.Del(store.prefix + path)
	return err
}

func (store s3store) GetReader(path string, nonce [32]byte) (rc io.ReadCloser, err error) {
	reader, err := store.bucket.GetReader(store.prefix + path)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func (store s3store) List(path string) ([]string, error) {
	list, err := store.bucket.List(store.prefix+path, "", "", 0)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	keys := []string{}
	for _, pic := range list.Contents {
		keys = append(keys, pic.Key)
	}
	return keys, nil
}

func (store s3store) PostReader(path string, file *bytes.Buffer, length int64, contentType string, nonce [32]byte) error {
	//reader := bytes.NewReader(file)
	// fmt.Println("POST READER NOW " + path)
	err := store.bucket.PutReader(store.prefix+path, file, length, contentType, s3.Private, s3.Options{})
	if err != nil {
		fmt.Println("error posting file: " + err.Error())
		return err
	}
	return nil
}

func (store s3store) GenNonce() ([32]byte, error) {
	var nonce [32]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		fmt.Printf("Failed to read random data: %v", err) // add error handling
		return nonce, err
	}
	return nonce, nil
}
