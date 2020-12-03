package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/rs/cors"
	"golang.org/x/crypto/blake2b"

	"github.com/stakwork/sphinx-meme/auth"
	"github.com/stakwork/sphinx-meme/ecdsa"
	"github.com/stakwork/sphinx-meme/frontend"
	"github.com/stakwork/sphinx-meme/ldat"
	"github.com/stakwork/sphinx-meme/storage"
)

// InitRouter creates the chi routes
func initRouter() *chi.Mux {
	r := initChi()

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode("pong")
	})

	r.Group(func(r chi.Router) {
		r.Get("/ask", ask)
		r.Post("/verify", verify)
		r.Get("/search/{searchTerm}", search) // do not return total_sats or total_buys
	})

	r.Group(func(r chi.Router) {
		r.Get("/podcast", getPodcast)
	})

	r.Group(func(r chi.Router) {
		r.Get("/", frontend.IndexRoute)
		r.Get("/static/*", frontend.StaticRoute)
		r.Get("/manifest.json", frontend.ManifestRoute)

		r.Get("/public/{muid}", getPublicMedia)
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.Verifier(auth.TokenAuth))
		r.Use(jwtauth.Authenticator)
		r.Use(auth.HostContext)
		r.Use(auth.PubKeyContext)

		r.Get("/mymedia", getMyMedia)              // only owner
		r.Get("/mymedia/{muid}", getMyMediaByMUID) // only owner
		r.Get("/media/{muid}", getMediaByMUID)

		r.Post("/file", uploadEncryptedFile)
		r.Get("/file/{token}", getMedia)

		r.Post("/public", uploadPublic)

		r.Post("/template", uploadTemplate)
		r.Get("/template/{muid}", getTemplate)
		r.Get("/templates", getTemplates)

		r.Put("/purchase/{muid}", mediaPurchase) // from owners relay node to update stats (and check current price)
	})

	return r
}

func getPodcast(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	latest, err := getLatestEpisode(url)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(latest)
}

func getTemplate(w http.ResponseWriter, r *http.Request) {

	muid := chi.URLParam(r, "muid")

	media := DB.getMediaWithDimensionsByMuid(muid)

	nonceBytes, err := hex.DecodeString(media.Nonce)
	var nonce [32]byte
	if err == nil {
		copy(nonce[:], nonceBytes)
	}

	reader, err := storage.Store.GetReader(muid, nonce)
	if err != nil {
		fmt.Println(err)
		fmt.Println("File not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer reader.Close()

	fmt.Println("")
	contentDisposition := fmt.Sprintf("attachment; filename=%s", media.Filename)
	w.Header().Set("Content-Disposition", contentDisposition)
	w.Header().Set("Content-Type", media.Mime)
	w.Header().Set("Content-Length", strconv.Itoa(int(media.Size)))
	io.Copy(w, reader)
}

func getPublicMedia(w http.ResponseWriter, r *http.Request) {

	muid := chi.URLParam(r, "muid")

	media := DB.getMediaByMUID(muid)

	nonceBytes, err := hex.DecodeString(media.Nonce)
	var nonce [32]byte
	if err == nil {
		copy(nonce[:], nonceBytes)
	}

	thumb := r.URL.Query().Get("thumb")

	themuid := muid
	if thumb == "true" {
		themuid = muid + "_thumb"
	}

	fmt.Println(themuid)
	reader, err := storage.Store.GetReader(themuid, nonce)
	if err != nil {
		fmt.Println(err)
		fmt.Println("File not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer reader.Close()

	fmt.Println("")
	contentDisposition := fmt.Sprintf("attachment; filename=%s", media.Filename)
	w.Header().Set("Content-Disposition", contentDisposition)
	w.Header().Set("Content-Type", media.Mime)
	// w.Header().Set("Content-Length", strconv.Itoa(int(media.Size)))
	io.Copy(w, reader)
}

func getTemplates(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	// pubKey := ctx.Value(auth.ContextKey).(string)

	medias := DB.getTemplates()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(medias)
}

func getMyMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pubKey := ctx.Value(auth.ContextKey).(string)

	medias := DB.getMyMedia(pubKey)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(medias)
}

func search(w http.ResponseWriter, r *http.Request) {
	searchTerm := chi.URLParam(r, "searchTerm")
	medias := DB.searchMedia(searchTerm)

	ms := []Media{}
	for _, m := range medias { // hidden vals
		m.TotalBuys = 0
		m.TotalSats = 0
		ms = append(ms, m)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ms)
}

// sent from owners relay node
func mediaPurchase(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pubKey := ctx.Value(auth.ContextKey).(string)

	muid := chi.URLParam(r, "muid")

	media := DB.mediaPurchase(pubKey, muid)
	if media.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("Media not found")
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(media)
}

func getMyMediaByMUID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pubKey := ctx.Value(auth.ContextKey).(string)

	muid := chi.URLParam(r, "muid")

	media := DB.getMyMediaByMUID(pubKey, muid)
	if media.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("Media not found")
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(media)
}

func getMediaByMUID(w http.ResponseWriter, r *http.Request) {
	muid := chi.URLParam(r, "muid")

	media := DB.getMediaByMUID(muid)
	media.TotalSats = 0 // hide stats, as no owner check
	media.TotalBuys = 0
	if media.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("Media not found")
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(media)
}

func uploadEncryptedFile(w http.ResponseWriter, r *http.Request) {
	uploadFile(w, r, false, false)
}

func uploadTemplate(w http.ResponseWriter, r *http.Request) {
	uploadFile(w, r, true, false)
}

func uploadPublic(w http.ResponseWriter, r *http.Request) {
	uploadFile(w, r, false, true)
}

// UploadFile uploads a file of any type
// need a "name" of "file"
func uploadFile(w http.ResponseWriter, r *http.Request, measureDimensions bool, thumb bool) {
	ctx := r.Context()
	pubKey := ctx.Value(auth.ContextKey).(string)

	fmt.Println("File Upload ===> ")

	// max of 32 MB files?
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("File too big")
		return
	}
	// FormFile returns the first file for the given key `file`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		json.NewEncoder(w).Encode("Error Retrieving the File")
		return
	}
	defer file.Close()

	p := uploadParams{}
	decodeForm(r.Form, &p)
	if p.TTL == 0 { // default to one year
		p.TTL = 60 * 60 * 24 * 365
	}

	filename := handler.Filename
	if filename == "" {
		filename = "file"
	}
	contentType := "application/octet-stream"
	contentTypes := handler.Header["Content-Type"]
	if len(contentTypes) > 0 {
		contentType = contentTypes[0]
	}

	imageWidth := 0
	imageHeight := 0
	if measureDimensions {
		imageWidth, imageHeight = getImageDimension(file)
	}

	var buf bytes.Buffer
	hasher, _ := blake2b.New256(nil) // hash it
	length, err := io.Copy(&buf, io.TeeReader(file, hasher))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hash := hasher.Sum(nil)
	defer file.Close()

	nonce, _ := storage.Store.GenNonce()
	nonceString := hex.EncodeToString(nonce[:])
	now := time.Now()
	media := Media{
		ID:          base64.URLEncoding.EncodeToString(hash[:]),
		OwnerPubKey: pubKey,
		Name:        p.Name,
		Description: p.Description,
		Tags:        p.Tags,
		Size:        handler.Size,
		Filename:    filename,
		Mime:        contentType,
		Nonce:       nonceString,
		TTL:         p.TTL,
		Price:       p.Price,
		Created:     &now,
		Updated:     &now,
		TotalBuys:   0,
		TotalSats:   0,
		Width:       imageWidth,
		Height:      imageHeight,
	}
	fmt.Printf("MEDIA: %+v\n", media)

	created, err := DB.createMedia(media)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	if thumb {
		go uploadThumb(media.ID, nonce, ioutil.NopCloser(bytes.NewReader(buf.Bytes())))
	}

	path := media.ID
	go storage.Store.PostReader(path, &buf, length, contentType, nonce)

	fmt.Println(length)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(created)
}

func getMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	mypubkey, _ := ctx.Value(auth.ContextKey).(string)
	host, _ := ctx.Value(auth.ContextHost).(string)

	mediaToken := chi.URLParam(r, "token") // full string
	parsed, err := ldat.Parse(mediaToken)
	if err != nil {
		fmt.Println("Error parsing terms")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	fmt.Println(parsed.Terms)
	sig := parsed.Terms.Sig

	terms := parsed.Terms
	muid := terms.Muid
	if muid == "" || host == "" {
		fmt.Println("No MUID")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if host != terms.Host {
		fmt.Println("Wrong Host")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	exp := terms.Exp
	if exp < time.Now().Unix() {
		fmt.Println("Access Expired")
		w.WriteHeader(http.StatusGone)
		return
	}

	media := DB.getMediaByMUID(muid)
	if media.ID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if mypubkey != terms.BuyerPubKey { // pubkey must match terms pubkey
		fmt.Println("Wrong Buyer Pub Key")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	nonceBytes, err := hex.DecodeString(media.Nonce)
	var nonce [32]byte
	if err == nil {
		copy(nonce[:], nonceBytes)
	}

	// the following logic is for non-owners (owner dont need token)
	if media.OwnerPubKey != mypubkey {

		bytesToVerify := parsed.Bytes
		bytes64 := base64.URLEncoding.EncodeToString(bytesToVerify)

		pubKeyZ, valid, err := ecdsa.VerifyAndExtract(bytes64, sig)
		if !valid || err != nil {
			fmt.Println("Cant Verify")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if media.OwnerPubKey != pubKeyZ {
			fmt.Println("Invalid Signature")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	fmt.Printf("GET: %s\n", muid)
	reader, err := storage.Store.GetReader(muid, nonce)
	if err != nil {
		fmt.Println(err)
		fmt.Println("File not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer reader.Close()

	contentDisposition := fmt.Sprintf("attachment; filename=%s", media.Filename)
	w.Header().Set("Content-Disposition", contentDisposition)
	w.Header().Set("Content-Type", media.Mime)
	w.Header().Set("Content-Length", strconv.Itoa(int(media.Size)))
	io.Copy(w, reader)
}

// NOT USED yet:

func deleteFile(w http.ResponseWriter, r *http.Request) {

	err := storage.Store.Delete("/deleteme")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("deleteme")
}

func initChi() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-User", "authorization"},
		ExposedHeaders:   []string{"Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
		//Debug:            true,
	})
	r.Use(cors.Handler)
	r.Use(middleware.Timeout(60 * time.Second))
	return r
}
