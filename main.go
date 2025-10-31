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

var MessageDB []route.Message = []route.Message{
	{ Id: "1", Client: []string{"admin", "user"}, Data: []route.MessageContent{
		{ Id: "1", Message: "Hello, I am Ken Lee", From: "admin", Time: "12:00" },
		{ Id: "2", Message: "What can I help you with today?", From: "admin", Time: "12:01" },
	}},
	{ Id: "2", Client: []string{"admin", "user2"}, Data: []route.MessageContent{
		{ Id: "1", Message: "Hello, I am Ken Lee", From: "admin", Time: "12:00" },
		{ Id: "2", Message: "What can I help you with today?", From: "admin", Time: "12:01" },
		{ Id: "3", Message: "I am interested in the porsche 911. Can you provide more details?", From: "user2", Time: "12:02" },
	}},
}


func main() {
	// gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(cors.Default())
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	});

	route.Router(router, &UserDB)
	route.Router(router, &MessageDB)

	router.Static("/public", "./public")
	router.Run(":8080")
}