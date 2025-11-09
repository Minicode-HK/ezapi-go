package http

import (
    "github.com/gin-gonic/gin"
)

// Helper function to standardize success responses
func SendSuccess(c *gin.Context, data interface{}, extra ...interface{}) {
    response := gin.H{
        "success": true,
        "data":    data,
    }
    
    // If extra data is provided, merge it into the response
    if len(extra) > 0 {
        if extraInfo, ok := extra[0].(gin.H); ok {
            for key, value := range extraInfo {
                response[key] = value
            }
        }
    }

    c.JSON(200, response)
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

