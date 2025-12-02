package auth

import (
	"github.com/Minicode-HK/ezapi-go/core/http"

	"github.com/gin-gonic/gin"
)

func AuthenticateMiddleware(provider Provider) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := provider.Authenticate(c)
		if err != nil {
			http.SendError(c, 401, "Unauthorized: "+err.Error())
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

