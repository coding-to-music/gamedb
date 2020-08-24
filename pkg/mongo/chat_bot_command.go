package mongo

import (
	"time"

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

func GetChatBotCommandsRecent() (commands []ChatBotCommand, err error) {

	filter := bson.D{{"message", bson.M{"$not": bson.M{"$regex": "help"}}}}

	cur, ctx, err := Find(CollectionChatBotCommands, 0, 100, bson.D{{"_id", -1}}, filter, nil, nil)
	if err != nil {
		return commands, err
	}

	defer func() {
		err = cur.Close(ctx)
		if err != nil {
			log.ErrS(err)
		}
	}()

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
