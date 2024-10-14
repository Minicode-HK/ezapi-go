package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"simple_backend_go/route"
)

var UserDB []route.User = []route.User{
	{ Id: "1" },
	{ Id: "2" },
}


func main() {
	// gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(cors.Default())
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	});

	route.Router(router, &UserDB)

	router.Static("/public", "./public")
	router.Run(":8080")
}