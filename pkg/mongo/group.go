package mongo

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	. "go.mongodb.org/mongo-driver/bson"
)

var ErrInvalidGroupID = errors.New("invalid group id")

// { name: 'text', url: 'text', headline: 'text' }, { weights: { name: 3, url: 2, headline: 1 }}

type Group struct {
	ID64          string    `bson:"_id"` // Too big for int64
	ID            int       `bson:"id"`
	CreatedAt     time.Time `bson:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at"`
	Name          string    `bson:"name"`
	Abbr          string    `bson:"abbreviation"`
	URL           string    `bson:"url"`
	AppID         int       `bson:"app_id"`
	Headline      string    `bson:"headline"`
	Summary       string    `bson:"summary"`
	Icon          string    `bson:"icon"`
	Trending      int64     `bson:"trending"`
	Members       int       `bson:"members"`
	MembersInChat int       `bson:"members_in_chat"`
	MembersInGame int       `bson:"members_in_game"`
	MembersOnline int       `bson:"members_online"`
	Error         string    `bson:"error"`
	Type          string    `bson:"type"`
}

func (group Group) BSON() (ret interface{}) {

	if group.CreatedAt.IsZero() {
		group.CreatedAt = time.Now()
	}

	group.UpdatedAt = time.Now()

	return M{
		"_id":             group.ID64,
		"id":              group.ID,
		"created_at":      group.CreatedAt,
		"updated_at":      group.UpdatedAt,
		"name":            group.Name,
		"abbreviation":    group.Abbr,
		"url":             group.URL,
		"app_id":          group.AppID,
		"headline":        group.Headline,
		"summary":         group.Summary,
		"icon":            group.Icon,
		"trending":        group.Trending,
		"members":         group.Members,
		"members_in_chat": group.MembersInChat,
		"members_in_game": group.MembersInGame,
		"members_online":  group.MembersOnline,
		"error":           group.Error,
		"type":            group.Type,
	}
}

func (group *Group) SetID(id string) {

	if len(id) < 18 {
		i, err := strconv.ParseInt(id, 10, 32)
		if err == nil {
			group.ID = int(i)
		}
	}
}

func (group Group) GetPath() string {
	return helpers.GetGroupPath(group.ID64, group.Name)
}

func (group Group) GetType() string {
	return helpers.GetGroupType(group.Type)
}

func (group Group) IsOfficial() bool {
	return helpers.IsGroupOfficial(group.Type)
}

func (group Group) GetURL() string {
	return helpers.GetGroupLink(group.Type, group.URL)
}

func (group Group) GetName() string {
	return helpers.GetGroupName(group.Name, group.ID64)
}

func (group Group) GetIcon() string {
	return helpers.AvatarBase + group.Icon
}

// Don't cache, as we need updatedAt to be live for notifications etc
func GetGroup(id string) (group Group, err error) {

	if !helpers.IsValidGroupID(id) {
		return group, ErrInvalidGroupID
	}

	if len(id) == 18 {
		err = FindOne(CollectionGroups, D{{"_id", id}}, nil, nil, &group)
	} else {
		i, err := strconv.ParseInt(id, 10, 32)
		if err == nil {
			err = FindOne(CollectionGroups, D{{"id", i}}, nil, nil, &group)
		}
	}

	if group.ID64 == "" {
		return group, ErrNoDocuments
	}

	return group, err
}

func GetGroupsByID(ids []string, projection M) (groups []Group, err error) {

	if len(ids) < 1 {
		return groups, nil
	}

	chunks := helpers.ChunkStrings(ids, 100)

	var wg sync.WaitGroup

	for _, chunk := range chunks {

		wg.Add(1)
		go func(chunk []string) {

			defer wg.Done()

			var id64sBSON A
			var idsBSON A

			for _, groupID := range chunk {
				if len(groupID) == 18 {
					id64sBSON = append(id64sBSON, groupID)
				} else {
					i, err := strconv.Atoi(groupID)
					log.Err(err)
					idsBSON = append(idsBSON, i)
				}
			}

			var or = A{}

			if len(id64sBSON) > 0 {
				or = append(or, M{"_id": M{"$in": id64sBSON}})
			}

			if len(idsBSON) > 0 {
				or = append(or, M{"id": M{"$in": idsBSON}})
			}

			resp, err := getGroups(0, 0, nil, D{{"$or", or}}, projection)
			if err != nil {
				log.Err(err)
				return
			}

			groups = append(groups, resp...)

		}(chunk)
	}

	wg.Wait()

	return groups, err
}

func GetGroups(limit int64, offset int64, sort D, filter D, projection M) (groups []Group, err error) {

	return getGroups(offset, limit, sort, filter, projection)
}

func getGroups(offset int64, limit int64, sort D, filter D, projection M) (groups []Group, err error) {

	cur, ctx, err := Find(CollectionGroups, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return groups, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var group Group
		err := cur.Decode(&group)
		if err != nil {
			log.Err(err, group.ID)
		}
		groups = append(groups, group)
	}

	return groups, cur.Err()
}
