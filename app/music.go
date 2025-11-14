package app

import (
	"ezapi-go/ez"
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
	ez.New[Music](&MusicDB).
		ResetableDB([]Music{
			{Id: "1", Title: "童話", Artist: "光良", MusicSheetUrl: "https://example.com/music-sheets/fairy-town.pdf"},
			{Id: "2", Title: "前世", Artist: "ヨルシカ", MusicSheetUrl: "https://example.com/music-sheets/love-confession.pdf"},
		}).
		CRUDRoutes("/api/music")
}
