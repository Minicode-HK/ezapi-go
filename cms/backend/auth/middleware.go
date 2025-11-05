package auth

import (
    "github.com/gin-gonic/gin"
)

func Middleware(authService *AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Try to get token from cookie first
        token, err := c.Cookie("cms_token")
        if err != nil || token == "" {
            // Fallback to Authorization header
            token = c.GetHeader("Authorization")
        }
        
        if token == "" {
            c.JSON(401, gin.H{
                "success": false,
                "message": "Unauthorized - No token provided",
            })
            c.Abort()
            return
        }
        
        session, valid := authService.ValidateToken(token)
        if !valid {
            c.JSON(401, gin.H{
                "success": false,
                "message": "Invalid or expired token",
            })
            c.Abort()
            return
        }
        
        // Store username in context
        c.Set("username", session.Username)
        c.Next()
    }
}