package storage

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/minio/sio"
	"golang.org/x/crypto/hkdf"
)

type localStore struct {
	prefix    string
	masterKey [32]byte
}

// Local ...
var Local localStore

func (store localStore) Init() {
	dir := os.Getenv("LOCAL_DIR")
	if dir == "" {
		dir = "files"
	}
	key := os.Getenv("LOCAL_ENCRYPTION_KEY")
	eKey, err := hex.DecodeString(key)
	var keyBytes [32]byte
	if err == nil {
		copy(keyBytes[:], eKey)
		Local.masterKey = keyBytes
	}
	Local.prefix = dir
}

func (store localStore) GetReader(path string, nonce [32]byte) (rc io.ReadCloser, err error) {
	// fmt.Println(store.prefix + "/" + path)
	f, err := os.Open(store.prefix + "/" + path)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	store.decryptReader(buf, f, nonce)
	return ioutil.NopCloser(buf), nil
}

func (store localStore) Delete(path string) error {
	var err = os.Remove(store.prefix + "/" + path)
	if err != nil {
		return err
	}
	fmt.Println("==> done deleting file")
	return nil
}

func (store localStore) PostReader(path string, buf *bytes.Buffer, length int64, contentType string, nonce [32]byte) error {
	file, err := os.Create(store.prefix + "/" + path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()

	if err := store.encryptReader(file, buf, nonce); err != nil {
		fmt.Println(err)
	}
	return file.Sync()
}

// UNIMPLEMENTED
func (store localStore) List(path string) ([]string, error) {
	return []string{}, nil
}

func (store localStore) encryptReader(dest io.Writer, src io.Reader, nonce [32]byte) error {

	// derive an encryption key from the master key and the nonce
	var key [32]byte
	kdf := hkdf.New(sha256.New, store.masterKey[:], nonce[:], nil)
	if _, err := io.ReadFull(kdf, key[:]); err != nil {
		fmt.Printf("Failed to derive encryption key: %v", err) // add error handling
		return err
	}

	encrypted, err := sio.EncryptReader(src, sio.Config{Key: key[:]})
	if err != nil {
		fmt.Printf("Failed to encrypted reader: %v", err) // add error handling
		return err
	}

	// the encrypted io.Reader can be used like every other reader - e.g. for copying
	if _, err := io.Copy(dest, encrypted); err != nil {
		fmt.Printf("Failed to copy data: %v", err) // add error handling
		return err
	}

	return nil
}

func (store localStore) decryptReader(dest io.Writer, src io.Reader, nonce [32]byte) error {

	// derive the encryption key from the master key and the nonce
	var key [32]byte
	kdf := hkdf.New(sha256.New, store.masterKey[:], nonce[:], nil)
	if _, err := io.ReadFull(kdf, key[:]); err != nil {
		fmt.Printf("Failed to derive encryption key: %v", err) // add error handling
		return err
	}

	if _, err := sio.Decrypt(dest, src, sio.Config{Key: key[:]}); err != nil {
		if _, ok := err.(sio.Error); ok {
			fmt.Printf("Malformed encrypted data: %v", err) // add error handling - here we know that the data is malformed/not authentic.
			return err
		}
		fmt.Printf("Failed to decrypt data: %v", err) // add error handling
		return err
	}
	return nil
}

func (store localStore) GenNonce() ([32]byte, error) {
	var nonce [32]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		fmt.Printf("Failed to read random data: %v", err) // add error handling
		return nonce, err
	}
	return nonce, nil
}
