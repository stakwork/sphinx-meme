package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const BUCKET_NAME = "sphinx-memes"

type aws3store struct {
	client *s3.Client
	prefix string
}

var bucket aws3store

func (store aws3store) Init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("failed to load s3 SDK configuration, %v", err)
	}

	bucket.client = s3.NewFromConfig(cfg)
	bucket.prefix = ""
}

func (store aws3store) PostReader(path string, file *bytes.Buffer, length int64, contentType string, nonce [32]byte) error {
	//reader := bytes.NewReader(file)
	// fmt.Println("POST READER NOW " + path)
	bucket := "sphinx-memes"
	_, err := store.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &path,
		Body:   file,
	})
	if err != nil {
		return err
	}
	return nil
}

func (store aws3store) GetReader(path string, nonce [32]byte) (rc io.ReadCloser, err error) {
	bucket := "sphinx-memes"
	result, err := store.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &path,
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

// UNIMPLEMENTED
func (store aws3store) List(path string) ([]string, error) {
	return []string{}, nil
}

// UNIMPLEMENTED
func (store aws3store) Delete(path string) error {
	return nil
}

func (store aws3store) GenNonce() ([32]byte, error) {
	var nonce [32]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		fmt.Printf("Failed to read random data: %v", err) // add error handling
		return nonce, err
	}
	return nonce, nil
}
