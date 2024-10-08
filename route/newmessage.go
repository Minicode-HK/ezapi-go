package route

import (
	"github.com/gin-gonic/gin"
)

type NewMessageContent struct {
	Id string `json:"id"`
	From string `json:"from"`
	Message string `json:"message"`
	Unread int `json:"unread"`
}	

func NewMessageGet(db *[]Message) gin.HandlerFunc {
  	return func(c *gin.Context) {
		var newMessageContent []NewMessageContent
		for _, message := range *db {
		newMessageContent = append(newMessageContent, NewMessageContent{
			Id: message.Id,
			From: message.Client[0],
			Message: message.Data[len(message.Data)-1].Message,
			Unread: len(message.Data),
		})
		}
		c.JSON(200, newMessageContent)
	}
}


func NewMessageRouter(router *gin.Engine, db *[]Message) {
	router.GET("/newmessage", NewMessageGet(db))
}