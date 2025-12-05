package auth

import (
	"github.com/gin-gonic/gin"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role	 string `json:"role"`
}

type Provider interface {
	Authenticate(context *gin.Context) (*User, error)
}

var GlobalAuthProvider Provider

// you can specify excluded paths that do not require authentication
func SetupAuthProvider(router *gin.Engine, provider Provider, excludedPaths ...string) {
	excludedPaths = append(excludedPaths, "/login", "/logout",  "/dashboard", "/dashboard/*", "/cms/api/auth-status", "/ping")
	router.Use(func(c *gin.Context) {
		// Check if the current path is in the excluded paths
		for _, path := range excludedPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}

			// pattern matching like /module/:id
			if c.FullPath() == path {
				c.Next()
				return
			}

			// wildcard matching like /module/*
			if len(path) > 0 && path[len(path) - 1] == '*' {
				prefix := path[:len(path)-1]
				if len(c.Request.URL.Path) > len(prefix) && prefix == c.Request.URL.Path[:len(path)-1] {
					c.Next()
					return
				}
			}
		}

		// Apply authentication middleware
		user, err := provider.Authenticate(c)
		if err != nil {
			c.JSON(401, gin.H{"error": "Unauthorized: " + err.Error()})
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Next()
	})
	GlobalAuthProvider = provider
}

func SetGlobalAuthProvider(provider Provider) {
	GlobalAuthProvider = provider
}

func GetGlobalAuthProvider() Provider {
    return GlobalAuthProvider
}

