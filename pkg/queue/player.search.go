package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type PlayersSearchMessage struct {
	Player mongo.Player `json:"player"`
}

func (m PlayersSearchMessage) Queue() rabbit.QueueName {
	return QueuePlayersSearch
}

func appsPlayersHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := PlayersSearchMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		player := elasticsearch.Player{}
		player.ID = payload.Player.ID
		player.PersonaName = payload.Player.PersonaName
		player.PersonaNameRecent = []string{} // todo
		player.VanityURL = payload.Player.VanityURL
		player.Avatar = payload.Player.Avatar
		player.CountryCode = payload.Player.CountryCode
		player.StateCode = payload.Player.StateCode
		player.LastBan = payload.Player.LastBan.Unix()
		player.GameBans = payload.Player.NumberOfGameBans
		player.Achievements = payload.Player.AchievementCount
		player.Achievements100 = payload.Player.AchievementCount100
		player.Continent = payload.Player.ContinentCode
		player.VACBans = payload.Player.NumberOfVACBans
		player.Level = payload.Player.Level
		player.PlayTime = payload.Player.PlayTime
		player.Badges = payload.Player.BadgesCount
		player.Games = payload.Player.GamesCount
		player.Friends = payload.Player.FriendsCount
		player.Comments = payload.Player.CommentsCount

		err = elasticsearch.IndexPlayer(player)
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
