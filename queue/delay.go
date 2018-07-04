package queue

import (
	"encoding/json"
	"math"
	"time"

	"github.com/streadway/amqp"
)

func processDelay(msg amqp.Delivery) (ack bool, requeue bool, err error) {

	time.Sleep(time.Second) // Minimum 1 second wait

	delayMessage := RabbitMessageDelay{}

	err = json.Unmarshal(msg.Body, &delayMessage)
	if err != nil {
		return false, true, err
		return false, false, err
	}

	if delayMessage.EndTime.UnixNano() > time.Now().UnixNano() {

		// Re-delay
		delayMessage.IncrementAttempts()

		bytes, err := json.Marshal(delayMessage)
		if err != nil {
			return false, true, err
			return false, false, err
		}

		// todo, handle returns from Produce and processDelay
		Produce(delayMessage.Queue, bytes)

	} else {

		// Add to original queue
		Produce(delayMessage.Queue, []byte(delayMessage.Message))
	}
}

type RabbitMessageDelay struct {
	Attempt   int
	StartTime time.Time
	EndTime   time.Time
	Queue     string // The queue it came from
	Message   string
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
