package core

import (
    "github.com/gin-gonic/gin"
)

// Helper function to standardize success responses
func SendSuccess(c *gin.Context, data interface{}) {
    c.JSON(200, gin.H{
        "success": true,
        "data":    data,
    })
}

func SendError(c *gin.Context, code int, message string) {
    c.JSON(code, gin.H{
        "success": false,
        "message": message,
    })
}

func SendErrorWithDetails(c *gin.Context, code int, message string, details interface{}) {
    c.JSON(code, gin.H{
        "success": false,
        "message": message,
        "details": details,
    })
}

