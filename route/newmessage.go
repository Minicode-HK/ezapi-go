package route

type NewMessageContent struct {
	Id string `json:"id"`
	From string `json:"from"`
	Message string `json:"message"`
	Unread int `json:"unread"`
}	
