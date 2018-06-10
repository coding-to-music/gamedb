package queue

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/streadway/amqp"
)

func processPics(msg amqp.Delivery) (ack bool, requeue bool) {

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
	MetaDataOnly    bool                                `json:"MetaDataOnly"`
	ResponsePending bool                                `json:"ResponsePending"`
	UnknownPackages []interface{}                       `json:"UnknownPackages"`
	UnknownApps     []interface{}                       `json:"UnknownApps"`
	Apps            map[string]RabbitMessagePicsProduct `json:"Apps"`
	Packages        map[string]RabbitMessagePicsProduct `json:"Packages"`
	JobID struct {
		SequentialCount int    `json:"SequentialCount"`
		StartTime       string `json:"StartTime"`
		ProcessID       int    `json:"ProcessID"`
		BoxID           int    `json:"BoxID"`
		Value           int64  `json:"Value"`
	} `json:"JobID"`
}

type RabbitMessagePicsProduct struct {
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
