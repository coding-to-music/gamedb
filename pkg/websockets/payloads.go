package websockets

type PubSubBasePayload struct {
	Pages []WebsocketPage `json:"p"`
}

type PubSubIDPayload struct {
	PubSubBasePayload
	ID int `json:"id"`
}

type PubSubStringPayload struct {
	PubSubBasePayload
	String string `json:"id"`
}

type PubSubIDStringsPayload struct {
	PubSubBasePayload
	IDs []string `json:"id"`
}

type PubSubChangesPayload struct {
	PubSubBasePayload
	Data [][]interface{} `json:"d"`
}

type AdminPayload struct {
	TaskID string `json:"task_id"`
	Action string `json:"action"`
	Time   int64  `json:"time"`
}

type ChatPayload struct {
	I            float32 `json:"i"`
	AuthorID     string  `json:"author_id"`
	AuthorUser   string  `json:"author_user"`
	AuthorAvatar string  `json:"author_avatar"`
	Content      string  `json:"content"`
	Channel      string  `json:"channel"`
	Time         string  `json:"timestamp"`
	Embeds       bool    `json:"embeds"`
}
