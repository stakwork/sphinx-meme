package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"

	"bytes"
	"encoding/base64"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/blake2b"

	"github.com/stakwork/sphinx-meme/storage"
)

func getImageDimension(file io.Reader) (int, int) {
	image, _, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Println(err)
		return 0, 0
	}
	return image.Width, image.Height
}

func goTest() {

	TTL := 60 * 60 * 24 * 365 * 100

	filename := "example10"

	contentType := "image/png"

	file, err := os.Open("templates/" + filename + ".png")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	var bufferRead bytes.Buffer
	img := io.TeeReader(file, &bufferRead)

	var buf bytes.Buffer
	hasher, _ := blake2b.New256(nil) // hash it
	length, err := io.Copy(&buf, io.TeeReader(img, hasher))
	if err != nil {
		fmt.Println(err)
		return
	}
	hash := hasher.Sum(nil)
	defer file.Close()

	// img must be read first, then its in bufferRead
	imageWidth, imageHeight := getImageDimension(&bufferRead)
	fmt.Println(imageWidth)

	nonce, _ := storage.Store.GenNonce()
	nonceString := hex.EncodeToString(nonce[:])
	now := time.Now()
	media := Media{
		ID:          base64.URLEncoding.EncodeToString(hash[:]),
		OwnerPubKey: "a",
		Name:        filename,
		Description: "payment template",
		Tags:        []string{},
		Size:        length,
		Filename:    filename,
		Mime:        contentType,
		Nonce:       nonceString,
		TTL:         int64(TTL),
		Price:       0,
		Created:     &now,
		Updated:     &now,
		TotalBuys:   0,
		TotalSats:   0,
		Width:       imageWidth,
		Height:      imageHeight,
		Template:    true,
	}
	fmt.Printf("MEDIA: %+v\n", media)

	created, err := DB.createMedia(media)
	if err != nil {
		fmt.Println("error", err)
		return
	}

	path := media.ID
	storage.Store.PostReader(path, &buf, length, contentType, nonce)

	fmt.Println(created)
}
