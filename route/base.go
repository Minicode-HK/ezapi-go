package route

import (
	"reflect"

	"github.com/gin-gonic/gin"
)

func Get[T any](inMemoryDB *[]T) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, inMemoryDB)
	}
}

func Post[T any](inMemoryDB *[]T) gin.HandlerFunc {
	return func(c *gin.Context) {
		var newMessage T
		c.BindJSON(&newMessage)
		*inMemoryDB = append(*inMemoryDB, newMessage)
		c.JSON(200, newMessage)
	}
}

func Delete[T any](inMemoryDB *[]T) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		for i, Message := range *inMemoryDB {
			// make sure the struct has an Id field
			var idField reflect.Value = reflect.ValueOf(Message).FieldByName("Id")
			if idField.IsValid() {
				if idField.String() == id {
					*inMemoryDB = append((*inMemoryDB)[:i], (*inMemoryDB)[i+1:]...)
					c.JSON(200, gin.H{"message": "Message deleted"})
					return
				}
			}
		}
		c.JSON(404, gin.H{"message": "Message not found"})
	}
}

func Put[T any](inMemoryDB *[]T) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var newMessage T
		c.BindJSON(&newMessage)
		for i, Message := range *inMemoryDB {
			// make sure the struct has an Id field
			var idField reflect.Value = reflect.ValueOf(Message).FieldByName("Id")
			if idField.IsValid() {
				if idField.String() == id {
					(*inMemoryDB)[i] = newMessage
					c.JSON(200, newMessage)
					return
				}
			}
		}
		c.JSON(404, gin.H{"message": "Message not found"})
	}
}

func Router[T any](router *gin.Engine, inMemoryDB *[]T) *gin.Engine {
	r := []rune(reflect.TypeOf(*inMemoryDB).Elem().Name())
	r[0] = r[0] + 32
	var moduleName string = string(r)
	router.GET("/" + moduleName, Get(inMemoryDB))
	router.POST("/" + moduleName, Post(inMemoryDB))
	router.DELETE("/" + moduleName + "/:id", Delete(inMemoryDB))
	router.PUT("/" + moduleName + "/:id", Put(inMemoryDB))
	return router
}

