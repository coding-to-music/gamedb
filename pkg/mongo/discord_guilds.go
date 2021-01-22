package mongo

import (
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

type DiscordGuild struct {
	ID   string `bson:"_id"`
	Name string `bson:"name"`
	Icon string `bson:"icon"`
}

func (guild DiscordGuild) BSON() bson.D {

	return bson.D{
		{"_id", guild.ID},
		{"name", guild.Name},
		{"icon", guild.Icon},
	}
}

func GetGuilds(guildIDs []string) (guilds map[string]DiscordGuild, err error) {

	guildIDs = helpers.UniqueString(guildIDs)

	a := bson.A{}
	for _, v := range guildIDs {
		a = append(a, v)
	}

	var filter = bson.D{{"_id", bson.M{"$in": a}}}

	cur, ctx, err := find(CollectionDiscordGuilds, 0, 100, nil, filter, nil, nil)
	if err != nil {
		return guilds, err
	}

	defer closeCursor(cur, ctx)

	guilds = map[string]DiscordGuild{}
	for cur.Next(ctx) {

		var guild DiscordGuild
		err := cur.Decode(&guild)
		if err != nil {
			log.ErrS(err)
		} else {
			guilds[guild.ID] = guild
		}
	}

	var discord *discordgo.Session

	for _, v := range guildIDs {
		if _, ok := guilds[v]; !ok {

			if discord == nil {
				discord, err = discordgo.New("Bot " + config.C.DiscordChatBotToken) // Bot must be in guild for access
				if err != nil {
					log.ErrS(err)
					break
				}
			}

			var guild DiscordGuild
			resp, err := discord.Guild(v)
			if val, ok := err.(*discordgo.RESTError); ok && val.Message.Code == 50001 { // Missing Access
				guild = DiscordGuild{ID: v}
			} else if err != nil {
				log.ErrS(err)
				continue
			} else {
				guild = DiscordGuild{ID: v, Name: resp.Name, Icon: resp.Icon}
			}

			guilds[v] = guild

			_, err = ReplaceOne(CollectionDiscordGuilds, bson.D{{"_id", v}}, guild)
			if err != nil {
				log.ErrS(err)
				break
			}
		}
	}

	return guilds, nil
}
