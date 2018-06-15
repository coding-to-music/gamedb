package websockets

type ChangeWebsocketPayload struct {
	ID            int                             `json:"id"`
	CreatedAtUnix int64                           `json:"created_at"`
	CreatedAtNice string                          `json:"created_at_nice"`
	Apps          []ChangeProductWebsocketPayload `json:"apps"`
	Packages      []ChangeProductWebsocketPayload `json:"packages"`
}

type ChangeProductWebsocketPayload struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
