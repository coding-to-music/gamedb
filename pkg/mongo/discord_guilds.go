package mongo

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
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
	UpdateAt time.Time `json:"update_at"`
}

func (guild DiscordGuild) BSON() bson.D {

	guild.UpdateAt = time.Now()

	if guild.Icon == "" {
		guild.Icon = "https://globalsteam.online/assets/img/discord2.png"
	}

	return bson.D{
		{"_id", guild.ID},
		{"name", guild.Name},
		{"icon", guild.Icon},
		{"members", guild.Members},
		{"update_at", guild.UpdateAt},
	}
}

func (guild *DiscordGuild) Update() (err error) {

	discord, err := discordgo.New("Bot " + config.C.DiscordChatBotToken)
	if err != nil {
		return err
	}

	resp, err := discord.GuildPreview(guild.ID)
	if err != nil {
		return err
	}

	if resp.ApproximateMemberCount == 0 {
		return
	}

	fullGuild := discordgo.Guild{Icon: resp.Icon} // todo, remove when pr gets merged in

	guild.ID = resp.ID
	guild.Name = resp.Name
	guild.Icon = fullGuild.IconURL()
	guild.Members = resp.ApproximateMemberCount

	_, err = ReplaceOne(CollectionDiscordGuilds, bson.D{{"_id", resp.ID}}, guild)
	return err
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
