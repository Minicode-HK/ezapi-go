package backend

import (
	"io"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupSystemStats(router *gin.Engine) {
	router.GET("/cms/api/stream/stats", func(c *gin.Context) {
		var m runtime.MemStats
		c.Stream(func(w io.Writer) bool {
			runtime.ReadMemStats(&m)
			c.SSEvent("stats", gin.H{
				"alloc":       m.Alloc,       
				"total_alloc": m.TotalAlloc,  
				"sys":         m.Sys,         
				"num_gc":      m.NumGC,       
				"goroutines":  runtime.NumGoroutine(),
				"timestamp":   time.Now().Format("15:04:05"),
			})
			time.Sleep(1 * time.Second)
			return true
		})
	}) 
}