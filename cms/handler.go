package cms

import (
    "strings"

    "github.com/gin-gonic/gin"

    "simple_backend_go/cms/auth"
    "simple_backend_go/cms/handlers"
    "simple_backend_go/route"
    
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
    playgroundHandler := handlers.NewPlaygroundHandler() 
    mockDataHandler := handlers.NewMockDataHandler()
    apiLogger := handlers.GetAPILogger()

    // CMS routes group
    cms := router.Group("/cms")

    // no-cache middleware
    cms.Use(func (c *gin.Context) {
        c.Writer.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
        c.Writer.Header().Set("Pragma", "no-cache")
        c.Writer.Header().Set("Expires", "0")
        c.Next()
    })

    // Protected routes
    protected := cms.Group("")
    protected.Use(auth.Middleware(authService))
    {
       
        // Auth
        protected.POST("/logout", authHandler.Logout)
        protected.GET("/api/me", authHandler.Me)
        
        // Content pages
        protected.GET("/admin", contentHandler.ServeLayout)
        protected.GET("/api-logger", contentHandler.ServeAPILogger)
        protected.GET("/api/content/dashboard", contentHandler.ServeDashboard)
        protected.GET("/api/content/snapshot", contentHandler.ServeSnapshot)
        protected.GET("/api/content/module/:name", contentHandler.ServeModule)
        protected.GET("/api/content/system_routes", contentHandler.ServeSystemRoutes)
        protected.GET("/api/content/playground", contentHandler.ServePlayground)
        
        // Schema
        protected.GET("/api/schemas", func(c *gin.Context) {
            schemaHandler.GetSchemas(c.Writer, c.Request)
        })
        protected.GET("/api/schema/:name", schemaHandler.GetSchema)
        
        // Snapshots
        protected.GET("/api/snapshots/modules", snapshotHandler.GetModules)
        protected.GET("/api/snapshots", snapshotHandler.List)
        protected.POST("/api/snapshots/save", snapshotHandler.Save)
        protected.POST("/api/snapshots/load", snapshotHandler.Load)
        protected.DELETE("/api/snapshots/:filename", snapshotHandler.Delete)

        // API Playground
        protected.POST("/api/playground/execute", playgroundHandler.ExecuteRequest)

        // Mock Data
        protected.POST("/api/mockdata/generate", mockDataHandler.Generate)
        protected.GET("/api/mockdata/preview/:module", mockDataHandler.Preview)

        // API Logger
        protected.GET("/api/logger/logs", apiLogger.List)
        protected.GET("/api/logger/logs/:id", apiLogger.GetByID)
        protected.DELETE("/api/logger/logs", apiLogger.Clear)
        protected.POST("/api/logger/toggle", apiLogger.Toggle)
        protected.GET("/api/logger/stats", apiLogger.Stats)
        protected.GET("/api/logger/export", apiLogger.Export)


        // System Routes Listing
         protected.GET("/system_routes", func(c *gin.Context) {
            // Get all registered routes from Gin
            routes := router.Routes()
            var routeList []map[string]interface{}
            
            // Get module registry for additional info
            modules := route.GetModuleRegistry()
            moduleMap := make(map[string]string)
            for _, mod := range modules {
                moduleMap[mod.BasePath] = mod.TypeName.Name()
            }
            
            for _, r := range routes {
                routeType := "api"
                module := "CMS"
                description := "Registered route"
                
                // Determine route type
                if strings.HasPrefix(r.Path, "/cms") || strings.HasPrefix(r.Path, "/snapshots") {
                    routeType = "cms"
                }
                
                // Try to find module from path
                if routeType == "api" {
                    for basePath, moduleName := range moduleMap {
                        if strings.HasPrefix(r.Path, basePath) {
                            module = moduleName
                            break
                        }
                    }
                }
                
                // Add description for common routes
                if routeType == "cms" {
                    if strings.Contains(r.Path, "/login") {
                        description = "Authentication"
                    } else if strings.Contains(r.Path, "/admin") {
                        description = "Admin interface"
                    } else if strings.Contains(r.Path, "/snapshot") {
                        description = "Snapshot management"
                    } else if strings.Contains(r.Path, "/static") {
                        description = "Static files"
                    }
                }

                if routeType == "api" {
                    if strings.Contains(r.Path, "/reset") {
                        description = "Reset in-memory databases to initial state"
                    }
                    // GET basePath + / - Schema listing
                    // GET basePath + /:id - Get :module by ID
                    // POST basePath + / - Create :module
                    // PUT basePath + /:id - Update :module by ID
                    // DELETE basePath + /:id - Delete :module by ID
                    if strings.HasSuffix(r.Path, "/") && r.Method == "GET" {
                        description = "List all " + module
                    } else if strings.HasSuffix(r.Path, "/:id") && r.Method == "GET" {
                        description = "Get " + module + " by ID"
                    } else if strings.HasSuffix(r.Path, "/") && r.Method == "POST" {
                        description = "Create new " + module
                    } else if strings.HasSuffix(r.Path, "/:id") && r.Method == "PUT" {
                        description = "Update " + module + " by ID"
                    } else if strings.HasSuffix(r.Path, "/:id") && r.Method == "DELETE" {
                        description = "Delete " + module + " by ID"
                    }

                }
                
                routeList = append(routeList, map[string]interface{}{
                    "method":      r.Method,
                    "path":        r.Path,
                    "type":        routeType,
                    "module":      module,
                    "description": description,
                })
            }
            
            c.JSON(200, gin.H{
                "success": true,
                "data":    routeList,
            })
        })
            
    }

    // Public routes
    cms.POST("/login", authHandler.Login)
    cms.GET("/login", authHandler.ServeLoginPage)

    router.LoadHTMLGlob("cms/static/private/content/dynamic/*.html")
    // Serve static files
    router.Static("/cms/static", "./cms/static")
    router.Static("/snapshots", cfg.SnapshotDir)
}