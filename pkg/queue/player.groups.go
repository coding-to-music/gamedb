package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	steamHelper "github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type PlayersGroupsMessage struct {
	PlayerID          int64   `json:"player_id"`
	PlayerPersonaName string  `json:"player_persona_name"`
	PlayerAvatar      string  `json:"player_avatar"`
	SkipPlayerGroups  bool    `json:"skip_groups"`
	SkipGroupUpdate   bool    `json:"skip_group"`
	UserAgent         *string `json:"user_agent"`
}

func playersGroupsHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := PlayersGroupsMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		if payload.SkipPlayerGroups {
			message.Ack(false)
			continue
		}

		// Old groups
		oldGroupsSlice, err := mongo.GetPlayerGroups(payload.PlayerID, 0, 0, nil)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		oldGroupsMap := map[string]bool{}
		for _, v := range oldGroupsSlice {
			oldGroupsMap[v.GroupID] = true
		}

		// Get new groups
		newGroupsResponse, _, err := steamHelper.GetSteam().GetUserGroupList(payload.PlayerID)
		err = steamHelper.AllowSteamCodes(err, 403)
		if err != nil {
			steamHelper.LogSteamError(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
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
			if _, ok := oldGroupsMap[k]; !ok {
				toAddIDs = append(toAddIDs, k)
			}
		}

		// Find new groups
		toAdd, err := mongo.GetGroupsByID(toAddIDs, nil)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		// Add
		var newPlayerGroupSlice []mongo.PlayerGroup
		for _, group := range toAdd {

			var name = helpers.TruncateString(group.Name, 1000, "") // Truncated as caused mongo driver issue

			newPlayerGroupSlice = append(newPlayerGroupSlice, mongo.PlayerGroup{
				PlayerID:     payload.PlayerID,
				PlayerName:   payload.PlayerPersonaName,
				PlayerAvatar: payload.PlayerAvatar,
				GroupID:      group.ID,
				GroupName:    name,
				GroupIcon:    group.Icon,
				GroupMembers: group.Members,
				GroupType:    group.Type,
				GroupURL:     group.URL,
			})
		}

		err = mongo.InsertPlayerGroups(newPlayerGroupSlice)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		// Delete
		var toDeleteIDs []string
		for _, v := range oldGroupsSlice {
			if _, ok := newGroupsMap[v.GroupID]; !ok {
				toDeleteIDs = append(toDeleteIDs, v.GroupID)
			}
		}

		err = mongo.DeletePlayerGroups(payload.PlayerID, toDeleteIDs)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		// Queue groups for update
		if !payload.SkipGroupUpdate {
			for id := range newGroupsMap {
				err = ProduceGroup(GroupMessage{ID: id, UserAgent: payload.UserAgent})
				err = helpers.IgnoreErrors(err, memcache.ErrInQueue, ErrIsBot)
				if err != nil {
					log.Err(err)
				}
			}
		}

		// Update player row
		var update = bson.D{
			{"groups_count", len(newGroupsMap)},
		}

		_, err = mongo.UpdateOne(mongo.CollectionPlayers, bson.D{{"_id", payload.PlayerID}}, update)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		// Clear caches
		var items = []string{
			memcache.MemcachePlayer(payload.PlayerID).Key,
		}

		err = memcache.Delete(items...)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
