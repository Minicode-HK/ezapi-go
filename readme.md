# Ezapi Golang version

Similar to this repo [ezapi](https://github.com/Minicode-HK/ezapi)

Build a simple RESTful API with `type` and `data`. The API does not require a database and provides endpoints for basic CRUD operations. All the data is stored in runtime memory.

Anytime you restart the server, the data will be reset. This is a good tool for testing and prototyping.

## Issue ( I guess?)

- I use reflection to access `Id` field in the struct.

## Example

Define the data structure in `route/message.go`

```golang
package route

type MessageContent struct {
 Id string `json:"id"`
 Message string `json:"message"`
 From string `json:"from"`
 Time string `json:"time"`
}

type Message struct {
 Id string `json:"id"`
 Client []string `json:"client"`
 Data []MessageContent `json:"data"`
}
```

Define the data in `main.go`

```golang
var MessageDB []route.Message = []route.Message{
 { Id: "1", Client: []string{"admin", "user"}, Data: []route.MessageContent{
  { Id: "1", Message: "Hello, I am Ken", From: "admin", Time: "12:00" },
  { Id: "2", Message: "What can I help you with today?", From: "admin", Time: "12:01" },
 }},
 { Id: "2", Client: []string{"admin", "user2"}, Data: []route.MessageContent{
  { Id: "1", Message: "Hello, I am Ken Lee", From: "admin", Time: "12:00" },
  { Id: "2", Message: "What can I help you with today?", From: "admin", Time: "12:01" },
  { Id: "3", Message: "I am interested in the porsche 911. Can you provide more details?", From: "user2", Time: "12:02" },
 }},
}
```

Register the route in `main.go`

```golang
func main() {
    router := gin.Default()
    // ...
    route.Router(router, &MessageDB)
   //  ...
}
```

---

Custom handler

Define handler in `route/newmessage.go`

```golang
package route

func NewmessageRouter(router *gin.Engine, db *[]Message) {
	router.GET("/newmessage", func(c *gin.Context) {
		c.JSON(200, /* you can access Message db here */)
	})
}
```

Register the route in `main.go`

```golang
func main() {
    router := gin.Default()
    // ...
    route.Router(router, &MessageDB)
    route.NewmessageRouter(router, &MessageDB)
   //  ...
}
```
