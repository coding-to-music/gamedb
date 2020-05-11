package queue

import (
	"github.com/Jleagle/rabbit-go"
)

type PlayersSearchMessage struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Flag        string `json:"flag"`
	Level       int    `json:"level"`
	Badges      int    `json:"badges"`
	Games       int    `json:"games"`
	Time        int    `json:"time"`
	BansGame    int    `json:"bans_game"`
	BansVAC     int    `json:"bans_vac"`
	BansVACLast int64  `json:"bans_vac_last"`
	Friends     int    `json:"friends"`
	Comments    int    `json:"comments"`
}

func appsPlayersHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		message.Ack(false)
	}
}
