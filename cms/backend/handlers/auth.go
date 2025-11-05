package handlers

import (
    "ezapi-go/cms/backend/auth"
    "ezapi-go/cms/backend/models"
    
    "github.com/gin-gonic/gin"
)

type AuthHandler struct {
    authService *auth.AuthService
}

func NewAuthHandler(authService *auth.AuthService) *AuthHandler {
    return &AuthHandler{
        authService: authService,
    }
}

func (h *AuthHandler) ServeLoginPage(c *gin.Context) {
    c.File("./cms/frontend/private/login.html")
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req struct {
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, models.APIResponse{
            Success: false,
            Message: "Invalid request",
        })
        return
    }
    
    token, err := h.authService.Login(req.Username, req.Password)
    if err != nil {
        c.JSON(401, models.APIResponse{
            Success: false,
            Message: "Invalid credentials",
        })
        return
    }
    
    // Set cookie
    c.SetCookie("cms_token", token, 86400, "/", "", false, true)
    
    c.JSON(200, models.APIResponse{
        Success: true,
        Data: gin.H{
            "token":    token,
            "username": req.Username,
        },
    })
}

func (h *AuthHandler) Logout(c *gin.Context) {
    token, _ := c.Cookie("cms_token")
    if token != "" {
        h.authService.Logout(token)
    }
    
    c.SetCookie("cms_token", "", -1, "/", "", false, true)
    
    c.JSON(200, models.APIResponse{
        Success: true,
        Message: "Logged out successfully",
    })
}

func (h *AuthHandler) Me(c *gin.Context) {
    username, _ := c.Get("username")
    
    c.JSON(200, models.APIResponse{
        Success: true,
        Data: gin.H{
            "username": username,
        },
    })
}