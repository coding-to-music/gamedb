package mongo

import (
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

type ChatBotCommand struct {
	Guild   string `bson:"guild"`
	Channel string `bson:"channel"`
	Author  string `bson:"author"`
	Message string `bson:"message"`
}

func (command ChatBotCommand) BSON() bson.D {

	return bson.D{
		{"guild", command.Guild},
		{"channel", command.Channel},
		{"author", command.Author},
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
