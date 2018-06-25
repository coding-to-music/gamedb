package queue

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/streadway/amqp"
)

func processDelay(msg amqp.Delivery) (ack bool, requeue bool, err error) {

	time.Sleep(time.Second) // Minimum 1 second wait

	queue := getQueueFromBytes(msg.Body)

	retrySTuff, err := getRetryStuffFromMessageBytes(queue, msg.Body)
	if err != nil {
		return false, true, err
		return false, false, err
	}

	if retrySTuff.EndTime.Unix() > time.Now().Unix() {

		// Re-delay

	} else {

		// Add to original queue

	}

	Produce(ProduceOptions{QueueApps, ), 1})
}

func getQueueFromBytes(data []byte) (string) {

	var localStr string
	err := json.NewDecoder(bytes.NewReader(data)).Decode(&struct {
		String *string `json:"jsonParam"`
	}{&localStr})
	if err != nil {
		log.Fatal(err)
	}

	return localStr
}

func getRetryStuffFromMessageBytes(queue string, bytes []byte) (stuff RabbitMessageDelay, err error) {

	switch queue {
	case QueueAppsData, QueuePackagesData:
		obj := RabbitMessageProduct{}
		err = json.Unmarshal(bytes, &obj)
		return obj.Retry, err
	case QueueChanges:
		obj := RabbitMessageChanges{}
		err = json.Unmarshal(bytes, &obj)
		return obj.Retry, err
	case QueuePlayers:
		obj := RabbitMessagePlayer{}
		err = json.Unmarshal(bytes, &obj)
		return obj.Retry, err
	case QueueDelays:
		obj := RabbitMessageDelay{}
		err = json.Unmarshal(bytes, &obj)
		return obj, err
	}

	return stuff, errors.New("unrecognised queue")
}

func requeueMessage(msg amqp.Delivery, queue string) error {

	delayMessage := RabbitMessageDelay{}
	delayMessage.Try = 1
	delayMessage.StartTime = time.Now()
	delayMessage.SetEndFromTrys()
	delayMessage.Queue = queue

	data, err := json.Marshal(delayMessage)
	if err != nil {
		return err
	}

	Produce(queue, data)
}

type RabbitMessageDelay struct {
	Try       int
	StartTime time.Time
	EndTime   time.Time
	Queue     string
}

func (d *RabbitMessageDelay) IncrementTrys() {
	d.Try++
	d.SetEndFromTrys()
}

func (d *RabbitMessageDelay) SetEndFromTrys() {

	var min float64 = 1
	var max float64 = 600

	var seconds = math.Pow(1.3, float64(d.Try))
	var minmaxed = math.Min(min+seconds, max)
	var rounded = math.Round(minmaxed)

	d.EndTime = d.StartTime.Add(time.Second * time.Duration(rounded))
}
