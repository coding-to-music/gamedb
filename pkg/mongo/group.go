package mongo

import (
	"errors"
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const AvatarBase = "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/"

var ErrInvalidGroupID = errors.New("invalid group id")

// { name: 'text', url: 'text', headline: 'text' }, { weights: { name: 3, url: 2, headline: 1 }}

type Group struct {
	ID64          string    `bson:"_id"` // Too big for int64
	ID            int       `bson:"id"`
	CreatedAt     time.Time `bson:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at"`
	Name          string    `bson:"name"`
	URL           string    `bson:"url"`
	Headline      string    `bson:"headline"`
	Summary       string    `bson:"summary"`
	Icon          string    `bson:"icon"`
	Members       int       `bson:"members"`
	MembersInChat int       `bson:"members_in_chat"`
	MembersInGame int       `bson:"members_in_game"`
	MembersOnline int       `bson:"members_online"`
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
		"url":             group.URL,
		"headline":        group.Headline,
		"summary":         group.Summary,
		"icon":            group.Icon,
		"members":         group.Members,
		"members_in_chat": group.MembersInChat,
		"members_in_game": group.MembersInGame,
		"members_online":  group.MembersOnline,
		"type":            group.Type,
	}
}

func (group Group) OutputForJSON() (output []interface{}) {

	return []interface{}{
		group.ID64,
		group.Name,
		group.GetPath(),
		group.GetIcon(),
		group.Headline,
		group.Members,
		group.URL,
		group.Type,
		group.GetLink(),
	}
}

func (group Group) GetPath() string {
	return "/groups/" + group.ID64 + "/" + slug.Make(group.Name)
}

func (group Group) GetLink() string {
	return "https://steamcommunity.com/groups/" + group.URL + "?utm_source=" + config.Config.GameDBShortName.Get()
}

func (group Group) GetName() string {
	if group.Name == "" {
		return "Group " + group.ID64
	}
	return group.Name
}

func (group Group) GetIcon() string {
	return AvatarBase + group.Icon
}

func GetGroup(id string) (group Group, err error) {

	if !helpers.IsValidGroupID(id) {
		return group, ErrInvalidGroupID
	}

	var item = helpers.MemcacheGroup(id)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &group, func() (interface{}, error) {

		var group Group

		if len(id) == 18 {
			err = FindDocumentByKey(CollectionGroups, "_id", id, nil, &group)
		} else {
			i, err := strconv.ParseInt(id, 10, 32)
			if err == nil {
				err = FindDocumentByKey(CollectionGroups, "id", i, nil, &group)
			}
		}

		return group, err
	})

	return group, err
}

func GetGroupsByID(ids []string, projection M, sort D) (groups []Group, err error) {

	if len(ids) < 1 {
		return groups, nil
	}

	var id64sBSON A
	var idsBSON A

	for _, v := range ids {
		if len(v) == 18 {
			id64sBSON = append(id64sBSON, v)
		} else {
			idsBSON = append(idsBSON, v)
		}
	}

	var or = A{}

	if len(id64sBSON) > 0 {
		or = append(or, M{"_id": M{"$in": id64sBSON}})
	}

	if len(idsBSON) > 0 {
		or = append(or, M{"id": M{"$in": idsBSON}})
	}

	return getGroups(0, 0, sort, M{"$or": or}, projection)
}

func GetGroups(limit int64, offset int64, sort D, filter M, projection M) (groups []Group, err error) {

	return getGroups(offset, limit, sort, filter, projection)
}

func getGroups(offset int64, limit int64, sort D, filter interface{}, projection M) (groups []Group, err error) {

	if filter == nil {
		filter = M{}
	}

	client, ctx, err := getMongo()
	if err != nil {
		return groups, err
	}

	ops := options.Find()
	if offset > 0 {
		ops.SetSkip(offset)
	}
	if limit > 0 {
		ops.SetLimit(limit)
	}
	if sort != nil {
		ops.SetSort(sort)
	}

	if projection != nil {
		ops.SetProjection(projection)
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionGroups.String())
	cur, err := c.Find(ctx, filter, ops)
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
