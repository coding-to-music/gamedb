package queue2

import (
	"sync"
	"time"

	"github.com/streadway/amqp"
)

var consumeLock = new(sync.Mutex)
var produceLock = new(sync.Mutex)

type RabbitQueue string

func (rq RabbitQueue) String() string {
	return string(rq)
}

const (
	QueueApps         RabbitQueue = "2Steam_Apps"
	QueueAppsData     RabbitQueue = "2Steam_Apps_Data"
	QueueBundlesData  RabbitQueue = "2Steam_Bundles_Data"
	QueueChangesData  RabbitQueue = "2Steam_Changes_Data"
	QueueDelaysData   RabbitQueue = "2Steam_Delays_Data"
	QueuePackages     RabbitQueue = "2Steam_Packages"
	QueuePackagesData RabbitQueue = "2Steam_Packages_Data"
	QueueProfiles     RabbitQueue = "2Steam_Profiles"
	QueueProfilesData RabbitQueue = "2Steam_Profiles_Data"
)

type BaseQueueMessage struct {
	QueueHistory []string
	FirstSeen    time.Time
	Payload      interface{}
}

type consumerInterface interface {
	process(msg amqp.Delivery) (requeue bool, err error)
}

type messageInterface interface {
	getPayloadStruct() interface{}
}
