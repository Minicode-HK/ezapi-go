package auth

import (
	"github.com/Minicode-HK/ezapi-go/core"
	"github.com/Minicode-HK/ezapi-go/core/feature"
	"github.com/Minicode-HK/ezapi-go/core/http"

	"github.com/gin-gonic/gin"
)

var userDB []User

func init() {
	userDB = feature.ResetableDatabase(&userDB, []User{
		{ID: "1", Username: "superadmin", Password: "superadmin", Role: "admin"},
		{ID: "2", Username: "user", Password: "user", Role: "user"},
	})

	// currently only support login function
	core.RegisterRouterWith(func(router *gin.Engine) {
		router.POST("/login", func(c *gin.Context) {
			var loginData struct {
				Username string `json:"username" binding:"required" form:"username"`
				Password string `json:"password" binding:"required" form:"password"`
			}
			if err := c.ShouldBind(&loginData); err != nil {
				http.SendError(c, 400, "Bad Request: "+err.Error())
				return
			}

			var authenticatedUser *User = feature.NewQueryBuilder(userDB).Where("Username", "=", loginData.Username).
				Where("Password", "=", loginData.Password).
				First();

			if authenticatedUser == nil {
				http.SendError(c, 401, "Unauthorized: invalid credentials")
				return
			}

			// Generate JWT token
			provider := GetGlobalAuthProvider()
			if provider == nil {
				http.SendError(c, 500, "Internal Server Error: auth provider not set")
				return
			}

			if jwtProvider, ok := provider.(*JWTProvider); ok {
				token, err := jwtProvider.GenerateToken(authenticatedUser.ID, authenticatedUser.Role)
				if err != nil {
					http.SendError(c, 500, "Internal Server Error: failed to generate token")
					return
				}

				http.SendSuccess(c, gin.H{
					"token": token,
				})
			} else {
				http.SendError(c, 500, "Currently only JWTProvider is supported")
			}
		})
	})
	
}