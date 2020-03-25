package main

import "log"

// Category constants
const (
	CategoryNone    = 0
	CategoryStore   = 1
	CategoryImage   = 2
	CategoryVideo   = 3
	CategoryAudio   = 4
	CategoryPodcast = 5
	CategoryAlbum   = 6
	CategoryBook    = 7
	CategoryFarm    = 8
)

var categories = map[string]uint{
	"none":    CategoryNone,
	"store":   CategoryStore,
	"image":   CategoryImage,
	"video":   CategoryVideo,
	"audio":   CategoryAudio,
	"podcast": CategoryPodcast,
	"album":   CategoryAlbum,
	"book":    CategoryBook,
	"farm":    CategoryFarm,
}

var categoryCodes = map[uint]string{}

func makeCategoryCodes() {
	for key, value := range categories {
		categoryCodes[value] = key
	}
	if len(categories) != len(categoryCodes) {
		log.Fatal("duplicate category code")
	}
}
