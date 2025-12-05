package frontend

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed dashboard/*
//go:embed dashboard/views/*
var content embed.FS

// GetFileSystem returns the http.FileSystem for the embedded dist folder
func GetFileSystem() http.FileSystem {
    fsys, err := fs.Sub(content, "dashboard")
    if err != nil {
        panic(err)
    }
    return http.FS(fsys)
}


func RegisterCMSFrontend(router *gin.Engine) {
    fileServer := GetFileSystem()

    router.GET("/dashboard", func(c *gin.Context) {
        c.FileFromFS("main.html", fileServer)
    })

	router.GET("/dashboard/*filepath", func(c *gin.Context) {
		c.FileFromFS(c.Param("filepath"), fileServer)
	})
}