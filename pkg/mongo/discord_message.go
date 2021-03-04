package mongo

import (
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

type DiscordMessage struct {
	GuildID      string    `bson:"guild_id"`
	ChannelID    string    `bson:"channel_id"`
	AuthorID     string    `bson:"author_id"`
	CommandID    string    `bson:"command_id"`
	AuthorName   string    `bson:"author_name"`
	AuthorAvatar string    `bson:"author_avatar"`
	Message      string    `bson:"message"`
	Slash        bool      `bson:"slash"`
	Time         time.Time `bson:"time"`
}

func (c DiscordMessage) BSON() bson.D {

	c.Time = time.Now()

	return bson.D{
		{"guild_id", c.GuildID},
		{"channel_id", c.ChannelID},
		{"author_id", c.AuthorID},
		{"command_id", c.CommandID},
		{"author_name", c.AuthorName},
		{"author_avatar", c.AuthorAvatar},
		{"message", c.Message},
		{"slash", c.Slash},
		{"time", c.Time},
	}
}

func (c DiscordMessage) GetAvatar() string {

	if c.AuthorAvatar == "" {
		c.AuthorAvatar = "/assets/img/discord.png"
	}

	return c.AuthorAvatar
}

func (c DiscordMessage) GetTableRowJSON(guilds map[string]DiscordGuild) []interface{} {

	return []interface{}{
		c.AuthorID,                     // 0
		c.AuthorName,                   // 1
		c.GetAvatar(),                  // 2
		c.GetCommand(),                 // 3
		c.Time.Unix(),                  // 4
		c.Time.Format(helpers.DateSQL), // 5
		guilds[c.GuildID].Name,         // 6
	}
}

func (c DiscordMessage) GetCommand() string {

	if c.Slash {
		return "/" + c.Message
	}

	if strings.HasPrefix(c.Message, "!") {
		return strings.Replace(c.Message, "!", ".", 1)
	}

	return c.Message
}

func GetChatBotCommandsRecent() (commands []DiscordMessage, err error) {

	cur, ctx, err := find(CollectionChatBotCommands, 0, 100, nil, bson.D{{"_id", -1}}, nil, nil)
	if err != nil {
		return commands, err
	}

	defer closeCursor(cur, ctx)

	for cur.Next(ctx) {

		var command DiscordMessage
		err := cur.Decode(&command)
		if err != nil {
			log.ErrS(err)
		} else {
			commands = append(commands, command)
		}
	}

	return commands, cur.Err()
}
