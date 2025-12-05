package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	_ "github.com/Minicode-HK/ezapi-go/app" // Import to trigger init()
	"github.com/Minicode-HK/ezapi-go/cms/backend"
	"github.com/Minicode-HK/ezapi-go/cms/backend/handlers"
	"github.com/Minicode-HK/ezapi-go/ez"
)

func main() {
    // Set up router
    router := gin.Default()
    
    // Configure CORS
    router.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
    }))
    
    // Health check endpoint
    router.GET("/ping", func(c *gin.Context) {
        c.String(http.StatusOK, "pong")
    })

    router.Use(handlers.LoggingMiddleware())  // This logs EVERYTHING including /api/*

    // TODO: currently auth is globally enforced except for specified routes.
    //       However, I think many routes / modules / earily development does not need auth at all.
    //       Therefore, we need to think a better way to solve this type of problem
    ez.SetupAuthProvider(router, ez.NewJWTProvider("your-secret-key"))

    // Add more resource routers here
    ez.SetupAllRouters(router)

    // Register CMS routes
    backend.RegisterCMSRoutes(router)
    
    // Get port from env or use default
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Server starting on port %s", port)
    router.Run(":" + port)
}