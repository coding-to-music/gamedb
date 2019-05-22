package mongo

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const AvatarBase = "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/"

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
	}
}

func (group Group) GetPath() string {
	return "/groups/" + group.ID64 + "/" + slug.Make(group.Name)
}

func (group Group) GetName() string {
	return group.Name
}

func (group Group) GetIcon() string {
	return AvatarBase + group.Icon
}

func GetGroup(id string) (group Group, err error) {

	// if !IsValidPlayerID(id) {
	// 	return group, ErrInvalidPlayerID
	// }

	if len(id) == 18 {
		err = FindDocument(CollectionGroups, "_id", id, nil, &group)
	} else {
		i, err := strconv.ParseInt(id, 10, 32)
		if err == nil {
			err = FindDocument(CollectionGroups, "id", i, nil, &group)
		}
	}

	return group, err
}

func GetGroupsByID(ids []int64, projection M) (groups []Group, err error) {

	if len(ids) < 1 {
		return groups, nil
	}

	var idsBSON A
	for _, v := range ids {
		idsBSON = append(idsBSON, v)
	}

	return getGroups(0, 0, D{{"name", 1}}, M{"id": M{"$in": idsBSON}}, projection)
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
