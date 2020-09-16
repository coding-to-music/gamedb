package mongo

import (
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

type DelayQueueMessage struct {
	UUID      string           `bson:"_id"`
	CreatedAt time.Time        `bson:"created_at"`
	UpdatedAt time.Time        `bson:"updated_at"`
	Queue     rabbit.QueueName `bson:"queue"`
	Attempt   int              `bson:"attempt"`
	Message   string           `bson:"message"`
}

func (m DelayQueueMessage) BSON() bson.D {

	m.UpdatedAt = time.Now()

	return bson.D{
		{"_id", m.UUID},
		// {"created_at", m.CreatedAt}, // Saved on the create call only
		{"updated_at", m.UpdatedAt},
		{"queue", m.Queue},
		{"attempt", m.Attempt},
		{"message", m.Message},
	}
}

func CreateDelayQueueMessage(m *rabbit.Message) {

	message := DelayQueueMessage{}
	message.CreatedAt = m.FirstSeen()
	message.UpdatedAt = m.LastSeen()
	message.UUID = m.UUID()
	message.Queue = m.LastQueue()
	message.Attempt = m.Attempt()
	message.Message = string(m.Message.Body)

	filter := bson.D{{"_id", message.UUID}}
	insert := bson.D{{"created_at", message.CreatedAt}}

	_, err := UpdateOneWithInsert(CollectionDelayQueue, filter, message.BSON(), insert)
	if err != nil {
		log.ErrS(err)
	}
}

func GetDelayQueueMessages(offset int64, sort bson.D) (messages []DelayQueueMessage, err error) {

	cur, ctx, err := Find(CollectionDelayQueue, offset, 100, sort, nil, nil, nil)
	if err != nil {
		return messages, err
	}

	defer close(cur, ctx)

	for cur.Next(ctx) {

		var message DelayQueueMessage
		err := cur.Decode(&message)
		if err != nil {
			log.ErrS(err)
		} else {
			messages = append(messages, message)
		}
	}

	return messages, cur.Err()
}
