package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/stakwork/sphinx-meme/auth"
	"github.com/stakwork/sphinx-meme/storage"
)

func main() {

	makeCategoryCodes()

	err := godotenv.Load()
	if err != nil {
		fmt.Println("no .env file")
	}

	initDB()
	auth.Init()
	storage.Init()
	r := initRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	fmt.Println("serving port " + port)
	http.ListenAndServe(":"+port, r)
}
