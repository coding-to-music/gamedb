package queue

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/streadway/amqp"
)

func processProduct(msg amqp.Delivery) (ack bool, requeue bool) {

	fmt.Println("x")

	// Get message payload
	message := new(RabbitMessagePics)

	err := json.Unmarshal(msg.Body, message)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(msg.Body))
		}

		return false, false
	}

	fmt.Println(string(msg.Body))

	return false, true
}

type RabbitMessagePics struct {
	ID           int                        `json:"ID"`
	ChangeNumber int                        `json:"ChangeNumber"`
	MissingToken bool                       `json:"MissingToken"`
	SHAHash      string                     `json:"SHAHash"`
	KeyValues    RabbitMessagePicsKeyValues `json:"KeyValues"`
	OnlyPublic   bool                       `json:"OnlyPublic"`
	UseHTTP      bool                       `json:"UseHttp"`
	HTTPURI      interface{}                `json:"HttpUri"`
}

type RabbitMessagePicsKeyValues struct {
	Name     string                       `json:"Name"`
	Value    interface{}                  `json:"Value"`
	Children []RabbitMessagePicsKeyValues `json:"Children"`
}
