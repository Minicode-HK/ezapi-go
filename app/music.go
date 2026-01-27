package app

import (
	"sync"

	"github.com/Minicode-HK/ezapi-go/ez"
	"github.com/gin-gonic/gin"
)


type Music struct {
	Id              string `json:"id"`
	Title           string `json:"title" binding:"required"`
	Artist          string `json:"artist" binding:"required"`
	MusicSheetUrl   string `json:"music_sheet_url"`
}

var MusicDB []Music


func init() {
	ez.New(&MusicDB).
		Seed([]Music{
			{Id: "1", Title: "童話", Artist: "光良", MusicSheetUrl: "https://example.com/music-sheets/fairy-town.pdf"},
			{Id: "2", Title: "前世", Artist: "ヨルシカ", MusicSheetUrl: "https://example.com/music-sheets/love-confession.pdf"},
			{Id: "3", Title: "Shape of You", Artist: "Ed Sheeran", MusicSheetUrl: "https://example.com/music-sheets/shape-of-you.pdf"},
		}).
		CRUD("/api/music").
		CustomRoutes(func(e *gin.Engine) {
			e.GET("/test", func(c *gin.Context) {

				var wg sync.WaitGroup
				wg.Add(3)
				// test the thread-safety 
				query := ez.Query(&MusicDB)
				go func() {
					defer wg.Done()
					// remove record with Id 1 with golang only
					ele := query.Where("Id", "=", "1").First()
					// remove ele from MusicDB
					if ele != nil {
						music := *ele
						MusicDB = append(MusicDB[:0], MusicDB[1:]...)
						_ = music  // to avoid unused variable warning
					}
				}()
				go func() {
					defer wg.Done()
					query.Where("Id", "=", "1").Delete()
				}()

				go func() {
					defer wg.Done()
					query.Add(&Music{Id: "4", Title: "New Song", Artist: "New Artist", MusicSheetUrl: "https://example.com/music-sheets/new-song.pdf"})
				}()
				wg.Wait()
				result1 := query.Get();

				ez.SendSuccess(c, gin.H{
					"result1": result1,
					"finalDB": MusicDB,
				})
			})
		})
}
