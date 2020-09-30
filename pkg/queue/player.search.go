package queue

import (
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.uber.org/zap"
)

type PlayersSearchMessage struct {
	Player   *mongo.Player `json:"player"`
	PlayerID int64         `json:"player_id"`
}

func (m PlayersSearchMessage) Queue() rabbit.QueueName {
	return QueuePlayersSearch
}

func appsPlayersHandler(message *rabbit.Message) {

	payload := PlayersSearchMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err("decode failed", zap.Error(err), zap.ByteString("message", message.Message.Body))
		sendToFailQueue(message)
		return
	}

	var mongoPlayer mongo.Player

	if payload.PlayerID > 0 {

		mongoPlayer, err = mongo.GetPlayer(payload.PlayerID)
		if err != nil {
			log.Err("retrieve player", zap.Error(err), zap.ByteString("message", message.Message.Body))
			sendToRetryQueue(message)
			return
		}

	} else if payload.Player != nil {

		mongoPlayer = *payload.Player

	} else {

		sendToFailQueue(message)
		return
	}

	player := elasticsearch.Player{}
	player.ID = mongoPlayer.ID
	player.PersonaName = mongoPlayer.PersonaName
	player.VanityURL = mongoPlayer.VanityURL
	player.Avatar = mongoPlayer.Avatar
	player.CountryCode = mongoPlayer.CountryCode
	player.StateCode = mongoPlayer.StateCode
	player.LastBan = mongoPlayer.LastBan.Unix()
	player.GameBans = mongoPlayer.NumberOfGameBans
	player.Achievements = mongoPlayer.AchievementCount
	player.Achievements100 = mongoPlayer.AchievementCount100
	player.Continent = mongoPlayer.ContinentCode
	player.VACBans = mongoPlayer.NumberOfVACBans
	player.Level = mongoPlayer.Level
	player.PlayTime = mongoPlayer.PlayTime
	player.Badges = mongoPlayer.BadgesCount
	player.BadgesFoil = mongoPlayer.BadgesFoilCount
	player.Games = mongoPlayer.GamesCount
	player.Friends = mongoPlayer.FriendsCount
	player.Comments = mongoPlayer.CommentsCount

	// Add aliases
	sixMonthsAgo := time.Now().AddDate(0, -6, 0).Unix()

	aliases, err := mongo.GetPlayerAliases(mongoPlayer.ID, 5, sixMonthsAgo)
	if err != nil {
		log.Err("retrieve aliases", zap.Error(err), zap.ByteString("message", message.Message.Body))
		sendToFailQueue(message)
		return
	}

	for _, v := range aliases {
		player.PersonaNameRecent = append(player.PersonaNameRecent, v.PlayerName)
	}

	// Send to Elastic
	err = elasticsearch.IndexPlayer(player)
	if err != nil {
		log.ErrS(err)
		sendToRetryQueue(message)
		return
	}

	message.Ack()
}
