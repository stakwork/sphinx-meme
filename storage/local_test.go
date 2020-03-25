package storage

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func check(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func TestLocalFile(t *testing.T) {
	godotenv.Load("../.env")
	Store := &Local
	Store.Init()

	nonce, _ := Store.GenNonce()

	contents := []byte("hello\n")

	b := bytes.NewBuffer(contents)

	err := Store.PostReader("hello.txt", b, 0, "", nonce)
	check(t, err)

	stream, err := Store.GetReader("hello.txt", nonce)
	defer stream.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)

	if string(buf.Bytes()) != string(contents) {
		t.Errorf("not equal")
	}
}

func TestLocalEncryption(t *testing.T) {
	godotenv.Load("../.env")
	Store := &Local
	Store.Init()

	nonce, _ := Store.GenNonce()

	os.Remove("files/encrypted.md")
	os.Remove("files/decrypted.md")

	f, err := os.Open("files/test.txt")
	check(t, err)
	defer f.Close()

	encrypted, err := os.Create("files/encrypted.txt")
	err = Store.encryptReader(encrypted, f, nonce)
	check(t, err)
	defer encrypted.Close()

	f2, err := os.Open("files/encrypted.txt")
	check(t, err)
	defer f2.Close()

	decrypted, err := os.Create("files/decrypted.txt")
	err = Store.decryptReader(decrypted, f2, nonce)
	check(t, err)
	defer decrypted.Close()

	contents, err := ioutil.ReadFile("files/test.txt")
	decryptedContents, err := ioutil.ReadFile("files/decrypted.txt")
	check(t, err)

	if string(contents) != string(decryptedContents) {
		t.Errorf("not equal")
	}
}
