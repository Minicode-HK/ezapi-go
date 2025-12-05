package backend

import (
	core_http "github.com/Minicode-HK/ezapi-go/core/http"
	"github.com/gin-gonic/gin"
)

func RegisterCMSBackend(router *gin.Engine) {
    group := router.Group("/cms")

    // Register API routes normally
    group.GET("/api/schema", func(c *gin.Context) {
        schema := GenerateSchema()
        core_http.SendSuccess(c, schema)
    })

    SetupSnapshots(router)
    SetupSystemStats(router)
}
