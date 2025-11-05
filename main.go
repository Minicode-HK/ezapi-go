package main

import (
    "log"
    "net/http"
    "os"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"

    "ezapi-go/core"
    "ezapi-go/cms/backend"
    "ezapi-go/cms/backend/handlers"
    _ "ezapi-go/app" // Import to trigger init()
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

    // Add more resource routers here
    core.SetupAllRouters(router)

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