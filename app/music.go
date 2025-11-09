package app

import (
    "ezapi-go/core"
    "ezapi-go/core/feature"
)


type Music struct {
	Id              string `json:"id"`
	Title           string `json:"title" binding:"required"`
	Artist          string `json:"artist" binding:"required"`
	MusicSheetUrl   string `json:"music_sheet_url"`
}

var MusicDB []Music

func GetMusicDB() *[]Music {
	return &MusicDB
}

func init() {
	
	// Initialize the in-memory database
	MusicDB = feature.ResetableDatabase(&MusicDB, []Music{
		{Id: "1", Title: "童話", Artist: "光良", MusicSheetUrl: "https://example.com/music-sheets/fairy-town.pdf"},
		{Id: "2", Title: "前世", Artist: "ヨルシカ", MusicSheetUrl: "https://example.com/music-sheets/love-confession.pdf"},
	})

	// Standard CRUD operations
	core.RegisterRouter(&MusicDB, "/api/music", &feature.IncrementalGenerator{})
}
