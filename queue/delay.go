package queue

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

type RabbitMessageDelay struct {
	OriginalQueue   string
	OriginalMessage string
}

func (d RabbitMessageDelay) getQueueName() string {
	return QueueDelaysData
}

func (d RabbitMessageDelay) getRetryData() RabbitMessageDelay {
	return RabbitMessageDelay{}
}

func (d RabbitMessageDelay) process(msg amqp.Delivery) (ack bool, requeue bool, err error) {

	//time.Sleep(time.Second) // Minimum 1 second wait

	if len(msg.Body) == 0 {
		return false, false, errEmptyMessage
	}

	delayMessage := RabbitMessageDelay{}

	err = json.Unmarshal(msg.Body, &delayMessage)
	if err != nil {
		return false, false, err
	}

	if len(delayMessage.OriginalMessage) == 0 {
		return false, false, errEmptyMessage
	}

	if delayMessage.EndTime.UnixNano() > time.Now().UnixNano() {

		// Re-delay
		fmt.Println("Re-delay: attemp: " + strconv.Itoa(delayMessage.Attempt))

		delayMessage.IncrementAttempts()

		bytes, err := json.Marshal(delayMessage)
		if err != nil {
			return false, false, err
		}

		err = Produce(delayMessage.getQueueName(), bytes)

	} else {

		// Add to original queue
		fmt.Println("Re-trying after attempt: " + strconv.Itoa(delayMessage.Attempt))

		err = Produce(delayMessage.getQueueName(), []byte(delayMessage.OriginalMessage))
	}

	if err != nil {
		return false, true, err
	}

	return true, false, nil
}

func (d *RabbitMessageDelay) IncrementAttempts() {
	d.Attempt++
	d.SetEndTime()
}

func (d *RabbitMessageDelay) SetEndTime() {

	var min float64 = 1
	var max float64 = 600

	var seconds = math.Pow(1.3, float64(d.Attempt))
	var minmaxed = math.Min(min+seconds, max)
	var rounded = math.Round(minmaxed)

	d.EndTime = d.StartTime.Add(time.Second * time.Duration(rounded))
}
