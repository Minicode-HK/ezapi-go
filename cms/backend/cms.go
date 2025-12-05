package backend

import (
	"github.com/Minicode-HK/ezapi-go/core/auth"
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

    // Auth status endpoint - tells frontend if auth is required
    group.GET("/api/auth-status", func(c *gin.Context) {
        authRequired := auth.GetGlobalAuthProvider() != nil
        core_http.SendSuccess(c, gin.H{
            "auth_required": authRequired,
        })
    })

    SetupSnapshots(router)
    SetupSystemStats(router)
}
