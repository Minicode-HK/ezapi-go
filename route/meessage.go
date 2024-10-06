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
