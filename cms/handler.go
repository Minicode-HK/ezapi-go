package cms

import (
    "simple_backend_go/cms/auth"
    "simple_backend_go/cms/handlers"
    
    "github.com/gin-gonic/gin"
)

func RegisterCMSRoutes(router *gin.Engine) {

    var cfg = DefaultConfig()

    // Initialize services
    authService := auth.NewAuthService(cfg.Users)

    // Initialize handlers
    authHandler := handlers.NewAuthHandler(authService)
    contentHandler := handlers.NewContentHandler()
    schemaHandler := handlers.NewSchemaHandler()
    snapshotHandler := handlers.NewSnapshotHandler(cfg.SnapshotDir)

    router.LoadHTMLGlob("cms/static/private/content/dynamic/*.html")
    
    // CMS routes group
    cms := router.Group("/cms")
    
    // Public routes
    cms.POST("/login", authHandler.Login)
    cms.GET("/login", authHandler.ServeLoginPage)
    
    // Protected routes
    protected := cms.Group("")
    protected.Use(auth.Middleware(authService))
    {
        // Auth
        protected.POST("/logout", authHandler.Logout)
        protected.GET("/api/me", authHandler.Me)
        
        // Content pages
        protected.GET("/admin", contentHandler.ServeLayout)
        protected.GET("/api/content/dashboard", contentHandler.ServeDashboard)
        protected.GET("/api/content/snapshot", contentHandler.ServeSnapshot)
        protected.GET("/api/content/module/:name", contentHandler.ServeModule)
        
        // Schema
        protected.GET("/api/schemas", func(c *gin.Context) {
            schemaHandler.GetSchemas(c.Writer, c.Request)
        })
        
        // Snapshots
        protected.GET("/api/snapshots/modules", snapshotHandler.GetModules)
        protected.GET("/api/snapshots", snapshotHandler.List)
        protected.POST("/api/snapshots/save", snapshotHandler.Save)
        protected.POST("/api/snapshots/load", snapshotHandler.Load)
        protected.DELETE("/api/snapshots/:filename", snapshotHandler.Delete)
    }
    
    // Serve static files
    router.Static("/cms/static", "./cms/static")
    router.Static("/snapshots", cfg.SnapshotDir)
}