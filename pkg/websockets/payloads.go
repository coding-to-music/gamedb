package websockets

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
