package mongo

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DiscordGuild struct {
	ID       string    `bson:"_id"`
	Name     string    `bson:"name"`
	Icon     string    `bson:"icon"`
	Members  int       `json:"members"`
	Requests int       `json:"requests"`
	UpdateAt time.Time `json:"update_at"`
}

func (guild DiscordGuild) BSON() bson.D {

	guild.UpdateAt = time.Now()

	if guild.Icon == "" {
		guild.Icon = "https://globalsteam.online/assets/img/discord2.png"
	}

	if !strings.HasPrefix(guild.Icon, "http") {
		dg := discordgo.Guild{ID: guild.ID, Icon: guild.Icon} // todo, remove when pr gets merged in
		guild.Icon = dg.IconURL()
	}

	requests, err := CountDocuments(CollectionChatBotCommands, bson.D{{"guild_id", guild.ID}}, 0)
	if err != nil {
		log.ErrS(err)
	}

	guild.Requests = int(requests)

	return bson.D{
		{"_id", guild.ID},
		{"name", guild.Name},
		{"icon", guild.Icon},
		{"members", guild.Members},
		{"requests", guild.Requests},
		{"update_at", guild.UpdateAt},
	}
}

func GetGuilds(offset int64, limit int64, sort bson.D, filter bson.D) (guilds []DiscordGuild, err error) {

	ops := options.Find()
	ops.SetCollation(&options.Collation{
		Locale:   "en",
		Strength: 2, // Case insensitive
	})

	cur, ctx, err := find(CollectionDiscordGuilds, offset, limit, filter, sort, nil, ops)
	if err != nil {
		return guilds, err
	}

	defer closeCursor(cur, ctx)

	for cur.Next(ctx) {

		var guild DiscordGuild
		err := cur.Decode(&guild)
		if err != nil {
			log.ErrS(err)
		} else {
			guilds = append(guilds, guild)
		}
	}

	return guilds, nil
}

func GetGuildsByIDs(ids []string) (guildsMap map[string]DiscordGuild, err error) {

	guildsMap = map[string]DiscordGuild{}

	a := bson.A{}
	for _, v := range helpers.UniqueString(ids) {
		a = append(a, v)
	}

	guilds, err := GetGuilds(0, 0, nil, bson.D{{"_id", bson.M{"$in": a}}})
	if err != nil {
		return guildsMap, err
	}

	guildsMap = map[string]DiscordGuild{}
	for _, v := range guilds {
		guildsMap[v.ID] = v
	}

	return guildsMap, nil
}
