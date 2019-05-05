package websockets

type PubSubBasePayload struct {
	Pages []WebsocketPage `json:"p"`
}

type PubSubIDPayload struct {
	PubSubBasePayload
	ID int `json:"id"`
}

type PubSubID64Payload struct {
	PubSubBasePayload
	ID int64 `json:"id"`
}

type PubSubChangesPayload struct {
	PubSubBasePayload
	Data [][]interface{} `json:"d"`
}

type AdminPayload struct {
	Message string `json:"message"`
}

type ChatPayload struct {
	AuthorID     string `json:"author_id"`
	AuthorUser   string `json:"author_user"`
	AuthorAvatar string `json:"author_avatar"`
	Content      string `json:"content"`
	Channel      string `json:"channel"`
}
