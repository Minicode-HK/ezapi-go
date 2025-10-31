package cms

import (
    "github.com/gin-gonic/gin"
)

func RegisterCMSRoutes(router *gin.Engine) {
    cms := router.Group("/cms")
    
    // Public routes (no auth required)
    cms.POST("/login", loginHandler)
    cms.GET("/login", serveLoginPage)
    
    // Protected routes (auth required)
    protected := cms.Group("")
    protected.Use(AuthMiddleware())
    {
        protected.GET("/admin", serveAdminPage)
        protected.GET("/api/schemas", func(c *gin.Context) {
            GetSchemaHandler(c.Writer, c.Request)
        })
        protected.POST("/logout", logoutHandler)
        protected.GET("/api/me", meHandler)
    }

    // Serve static files
    router.Static("/cms/static", "./cms/static")
}

func loginHandler(c *gin.Context) {
    var credentials struct {
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&credentials); err != nil {
        c.JSON(400, gin.H{"success": false, "message": "Invalid request"})
        return
    }
    
    token, err := authManager.Login(credentials.Username, credentials.Password)
    if err != nil {
        c.JSON(401, gin.H{"success": false, "message": "Invalid credentials"})
        return
    }
    
    // Set cookie with httpOnly=false so JavaScript can read it
    // In production, consider using httpOnly=true and rely only on cookie
    c.SetSameSite(3) // SameSiteStrictMode
	c.SetCookie("cms_token", token, 86400, "/", "", false, true)
    
    c.JSON(200, gin.H{
        "success": true,
        "user":    credentials.Username,
    })
}

func logoutHandler(c *gin.Context) {
    token := c.GetHeader("Authorization")
    if token == "" {
        token, _ = c.Cookie("cms_token")
    }
    
    if token != "" {
        authManager.Logout(token)
    }
    
    // Clear cookie
    c.SetSameSite(3)
    c.SetCookie("cms_token", "", -1, "/", "", false, false)
    c.JSON(200, gin.H{"success": true, "message": "Logged out"})
}

func meHandler(c *gin.Context) {
    username, _ := c.Get("username")
    c.JSON(200, gin.H{
        "success":  true,
        "username": username,
    })
}

func serveLoginPage(c *gin.Context) {
    c.File("./cms/static/admin/login.html")
}

func serveAdminPage(c *gin.Context) {
    c.File("./cms/static/admin/index.html")
}