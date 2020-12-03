package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"

	"github.com/stakwork/sphinx-meme/storage"
)

// MigrateThumbnails images
func MigrateThumbnails() {

	oy := DB.getAllMedia()
	fmt.Printf("=> %+v\n", len(oy))

	for i, m := range oy {
		if i > 6585 && m.Mime == "image/jpg" && m.Width == 0 && m.Height == 0 {
			fmt.Printf("=> %+v\n", i)
			tryMigrate(m.ID, m.Nonce)
		}
	}
}

func tryMigrate(muid string, nonceString string) error {
	// fmt.Printf("=> TRY MIGRATE %+v\n", "hi")
	nonceBytes, err := hex.DecodeString(nonceString)
	var nonce [32]byte
	if err != nil {
		return err
	}
	copy(nonce[:], nonceBytes)

	reader, err := storage.Store.GetReader(muid, nonce)
	if err != nil {
		return err
	}

	return uploadThumb(muid, nonce, reader)
}

// Axj5psD9cSYQWWwhDrgHj4EY5MotgrD79cYznantwzA=
// MWJtCOvFrD8wcNi3oW5uyNDr1aCk3MzmaLn7WYwCFhQ=
// W5ZqIyOo5c9k_ejyZKu3WDVdQ1cLFqxFn6MK9RFZO1A=
func uploadThumb(muid string, nonce [32]byte, reader io.ReadCloser) error {
	defer reader.Close()

	img, err := jpeg.Decode(reader)
	if err != nil {
		return err
	}

	fmt.Printf("=> uploadThumb %+v\n", muid)

	min := calcMin(img.Bounds().Max.X, img.Bounds().Max.Y)
	croppedImg, err := cutter.Crop(img, cutter.Config{
		Width: min, Height: min,
		Anchor: image.Point{1, 1},
		Mode:   cutter.TopLeft,
	})

	circleImg := image.NewRGBA(image.Rect(0, 0, min, min))
	draw.Draw(circleImg, circleImg.Bounds(), &image.Uniform{color.Transparent}, image.ZP, draw.Src) //white
	r := int(min / 2)
	p := image.Point{X: r, Y: r}
	draw.DrawMask(circleImg, circleImg.Bounds(), croppedImg, image.ZP, &circle{p, r}, image.ZP, draw.Over)

	thumb := resize.Thumbnail(60, 60, circleImg, resize.Lanczos3)

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, thumb, nil)
	if err != nil {
		should(err)
		return err
	}

	storage.Store.PostReader(muid+"_thumb", buf, int64(buf.Len()), "image/jpg", nonce)

	return nil
}

func calcMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type circle struct {
	p image.Point
	r int
}

func (c *circle) ColorModel() color.Model {
	return color.AlphaModel
}

func (c *circle) Bounds() image.Rectangle {
	return image.Rect(c.p.X-c.r, c.p.Y-c.r, c.p.X+c.r, c.p.Y+c.r)
}

func (c *circle) At(x, y int) color.Color {
	xx, yy, rr := float64(x-c.p.X)+0.5, float64(y-c.p.Y)+0.5, float64(c.r)
	if xx*xx+yy*yy < rr*rr {
		return color.Alpha{255}
	}
	return color.Alpha{0}
}

func getAndCropImage(path string) {
	url := "https://h2h.sfo2.digitaloceanspaces.com/" + path
	response, err := http.Get(url)
	if err == nil {
		defer response.Body.Close()
		bodyBytes, err := ioutil.ReadAll(response.Body)
		should(err)

		r2 := bytes.NewReader(bodyBytes)
		img, err := jpeg.Decode(r2)
		should(err)

		min := calcMin(img.Bounds().Max.X, img.Bounds().Max.Y)
		croppedImg, err := cutter.Crop(img, cutter.Config{
			Width: min, Height: min,
			Anchor: image.Point{1, 1},
			Mode:   cutter.TopLeft,
		})

		circleImg := image.NewRGBA(image.Rect(0, 0, min, min))
		draw.Draw(circleImg, circleImg.Bounds(), &image.Uniform{color.Transparent}, image.ZP, draw.Src) //white
		r := int(min / 2)
		p := image.Point{X: r, Y: r}
		draw.DrawMask(circleImg, circleImg.Bounds(), croppedImg, image.ZP, &circle{p, r}, image.ZP, draw.Over)

		thumb := resize.Thumbnail(60, 60, circleImg, resize.Lanczos3)

		// f, err := os.Create("img.png")
		// defer f.Close()
		// png.Encode(f, thumb)
		// if err != nil {
		// 	should(err)
		// 	return
		// }
		fmt.Println("PULOAD!")

		buf := new(bytes.Buffer)
		if err := png.Encode(buf, thumb); err != nil {
			fmt.Printf("unable to encode png %s\n", err.Error)
			return
		}
		// err = storage.Store.PutReader(path+"_thumb", buf, int64(len(buf.Bytes())), "image/png", s3.PublicRead)
		// if err != nil {
		// 	fmt.Printf("error uploading image: %s\n", err.Error())
		// }
	} else {
		fmt.Println(err)
	}
}

func should(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
