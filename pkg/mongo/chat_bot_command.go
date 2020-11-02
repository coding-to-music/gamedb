package mongo

import (
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

type ChatBotCommand struct {
	GuildID      string    `bson:"guild_id"`
	ChannelID    string    `bson:"channel_id"`
	AuthorID     string    `bson:"author_id"`
	AuthorName   string    `bson:"author_name"`
	AuthorAvatar string    `bson:"author_avatar"`
	CommandID    string    `bson:"command_id"`
	Message      string    `bson:"message"`
	Time         time.Time `bson:"time"`
}

func (command ChatBotCommand) BSON() bson.D {

	return bson.D{
		{"time", command.Time},
		{"guild_id", command.GuildID},
		{"channel_id", command.ChannelID},
		{"author_id", command.AuthorID},
		{"author_name", command.AuthorName},
		{"author_avatar", command.AuthorAvatar},
		{"message", command.Message},
	}
}

func (command ChatBotCommand) GetTableRowJSON(guilds map[string]DiscordGuild) []interface{} {

	return []interface{}{
		command.AuthorID,                     // 0
		command.AuthorName,                   // 1
		command.AuthorAvatar,                 // 2
		command.GetCommand(),                 // 3
		command.Time.Unix(),                  // 4
		command.Time.Format(helpers.DateSQL), // 5
		guilds[command.GuildID].Name,         // 6
	}
}

func (command ChatBotCommand) GetCommand() string {

	// Show all command prefixes as a full stop
	if strings.HasPrefix(command.Message, "!") {
		return strings.Replace(command.Message, "!", ".", 1)
	}

	return command.Message
}

func GetChatBotCommandsRecent() (commands []ChatBotCommand, err error) {

	filter := bson.D{{"message", bson.M{"$not": bson.M{"$regex": "help"}}}}

	cur, ctx, err := Find(CollectionChatBotCommands, 0, 100, bson.D{{"_id", -1}}, filter, nil, nil)
	if err != nil {
		return commands, err
	}

	defer closeCursor(cur, ctx)

	for cur.Next(ctx) {

		var command ChatBotCommand
		err := cur.Decode(&command)
		if err != nil {
			log.ErrS(err)
		} else {
			commands = append(commands, command)
		}
	}

	return commands, cur.Err()
}
