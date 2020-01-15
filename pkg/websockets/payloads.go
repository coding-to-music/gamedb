package websockets

type PubSubBasePayload struct {
	Pages []WebsocketPage `json:"p"`
}

type IntPayload struct {
	PubSubBasePayload
	ID int `json:"id"`
}

type StringPayload struct {
	PubSubBasePayload
	String string `json:"id"`
}

type StringsPayload struct {
	PubSubBasePayload
	IDs []string `json:"id"`
}

type ChangesPayload struct {
	PubSubBasePayload
	Data [][]interface{} `json:"d"`
}

type AdminPayload struct {
	PubSubBasePayload
	TaskID string `json:"task_id"`
	Action string `json:"action"`
	Time   int64  `json:"time"`
}

type ChatBotPayload struct {
	PubSubBasePayload
	AuthorID     string `json:"author_id"`
	AuthorName   string `json:"author_name"`
	AuthorAvatar string `json:"author_avatar"`
	Message      string `json:"message"`
}

type ChatPayload struct {
	PubSubBasePayload
	I            float32 `json:"i"`
	AuthorID     string  `json:"author_id"`
	AuthorUser   string  `json:"author_user"`
	AuthorAvatar string  `json:"author_avatar"`
	Content      string  `json:"content"`
	Channel      string  `json:"channel"`
	Time         string  `json:"timestamp"`
	Embeds       bool    `json:"embeds"`
}
