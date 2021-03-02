package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type PlayersGroupsMessage struct {
	Player                    mongo.Player `json:"player"`
	SkipGroupUpdate           bool         `json:"skip_group"`
	ForceResavingPlayerGroups bool         `json:"force_resaving_player_groups"`
	UserAgent                 *string      `json:"user_agent"`
}

func (m PlayersGroupsMessage) Queue() rabbit.QueueName {
	return QueuePlayersGroups
}

func playersGroupsHandler(message *rabbit.Message) {

	payload := PlayersGroupsMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	// Websocket
	defer sendPlayerWebsocket(payload.Player.ID, "group", message)

	if payload.Player.Removed {
		message.Ack()
		return
	}

	// Old groups
	oldGroupsSlice, err := mongo.GetPlayerGroups(payload.Player.ID, 0, 0, nil)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	oldGroupsMap := map[string]bool{}
	for _, v := range oldGroupsSlice {
		oldGroupsMap[v.GroupID] = true
	}

	// Get new groups
	newGroupsResponse, err := steam.GetSteam().GetUserGroupList(payload.Player.ID)

	if err == steamapi.ErrProfileMissing || err == steamapi.ErrProfilePrivate {
		message.Ack()
		return
	}

	err = steam.AllowSteamCodes(err)
	if err != nil {
		steam.LogSteamError(err, zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	newGroupsMap := map[string]bool{}
	for _, v := range newGroupsResponse.GetIDs() {
		v, err = helpers.IsValidGroupID(v)
		if err == nil {
			newGroupsMap[v] = true
		}
	}

	// Make list of groups to add
	var toAddIDs []string
	for k := range newGroupsMap {
		if _, ok := oldGroupsMap[k]; !ok || payload.ForceResavingPlayerGroups {
			toAddIDs = append(toAddIDs, k)
		}
	}

	if len(toAddIDs) > 0 {

		// Find new groups
		toAdd, err := mongo.GetGroupsByID(toAddIDs, nil)
		if err != nil {
			log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
			sendToRetryQueue(message)
			return
		}

		// Add
		var newPlayerGroupSlice []mongo.PlayerGroup
		for _, group := range toAdd {

			newPlayerGroupSlice = append(newPlayerGroupSlice, mongo.PlayerGroup{
				PlayerID:      payload.Player.ID,
				PlayerName:    payload.Player.PersonaName,
				PlayerAvatar:  payload.Player.Avatar,
				PlayerLevel:   payload.Player.Level,
				PlayerCountry: payload.Player.CountryCode,
				PlayerGames:   payload.Player.GamesCount,
				GroupID:       group.ID,
				GroupName:     helpers.TruncateString(group.Name, 1000, ""), // Truncated as caused mongo driver issue
				GroupIcon:     group.Icon,
				GroupMembers:  group.Members,
				GroupType:     group.Type,
				GroupURL:      group.URL,
			})
		}

		err = mongo.ReplacePlayerGroups(newPlayerGroupSlice)
		if err != nil {
			log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
			sendToRetryQueue(message)
			return
		}
	}

	// Delete
	var toDeleteIDs []string
	for _, v := range oldGroupsSlice {
		if _, ok := newGroupsMap[v.GroupID]; !ok {
			toDeleteIDs = append(toDeleteIDs, v.GroupID)
		}
	}

	err = mongo.DeletePlayerGroups(payload.Player.ID, toDeleteIDs)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Queue groups for update
	if !payload.SkipGroupUpdate {
		for id := range newGroupsMap {
			err = ProduceGroup(GroupMessage{ID: id, UserAgent: payload.UserAgent})
			err = helpers.IgnoreErrors(err, memcache.ErrInQueue, ErrIsBot)
			if err != nil {
				log.ErrS(err)
			}
		}
	}

	// Update player row
	var update = bson.D{
		{"groups_count", len(newGroupsMap)},
	}

	_, err = mongo.UpdateOne(mongo.CollectionPlayers, bson.D{{"_id", payload.Player.ID}}, update, nil)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Clear caches
	var items = []string{
		memcache.ItemPlayer(payload.Player.ID).Key,
	}

	err = memcache.Delete(items...)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Update Elastic
	err = ProducePlayerSearch(nil, payload.Player.ID)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	//
	message.Ack()
}
