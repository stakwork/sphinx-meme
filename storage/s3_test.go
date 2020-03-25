package storage

import (
	"testing"

	"github.com/joho/godotenv"
)

func TestS3(t *testing.T) {
	godotenv.Load("../.env")
	Store = &space
	Store.Init()

	// list, err := Store.List("x6uwdza6f3fm/")
	// if err != nil {
	// 	fmt.Println(err)
	// 	t.Fatalf(err.Error())
	// 	return
	// }
	// fmt.Printf("list %+v\n", list)
	// fileNames := []string{}
	// for _, file := range list.Contents {
	// 	idx := strings.LastIndex(file.Key, "/")
	// 	name := file.Key[idx+1:]
	// 	if len(name) > 0 {
	// 		fileNames = append(fileNames, name)
	// 	}
	// }
	// if len(fileNames) > 0 {
	// 	fmt.Printf("file names: %+v\n", fileNames)
	// }

	// gorrilla, err := Store.GetReader("x6uwdza6f3fm/gorrilla.jpg")
	// if err != nil {
	// 	fmt.Println(err)
	// 	t.Fatalf(err.Error())
	// 	return
	// }

	// fmt.Printf("gorrilla: %+v\n", gorrilla)

	if false {
		t.Fatalf("nope")
	}
}
