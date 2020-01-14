package mongo

import (
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

type ChatBotCommand struct {
	GuildID      string `bson:"guild_id"`
	ChannelID    string `bson:"channel_id"`
	AuthorID     string `bson:"author_id"`
	AuthorName   string `bson:"author_name"`
	AuthorAvatar string `bson:"author_avatar"`
	Message      string `bson:"message"`
}

func (command ChatBotCommand) BSON() bson.D {

	return bson.D{
		{"guild_id", command.GuildID},
		{"channel_id", command.ChannelID},
		{"author_id", command.AuthorID},
		{"author_name", command.AuthorName},
		{"author_avatar", command.AuthorAvatar},
		{"message", command.Message},
	}
}

func GetChatBotCommands() (commands []ChatBotCommand, err error) {

	cur, ctx, err := Find(CollectionChatBotCommands, 0, 100, bson.D{{"_id", -1}}, nil, nil, nil)
	if err != nil {
		return commands, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var command ChatBotCommand
		err := cur.Decode(&command)
		if err != nil {
			log.Err(err)
		} else {
			commands = append(commands, command)
		}
	}

	return commands, cur.Err()
}
