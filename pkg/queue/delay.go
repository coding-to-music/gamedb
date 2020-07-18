package queue

import (
	"math"
	"os"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/log"
)

func delayHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		time.Sleep(time.Second / 10)

		// writeToFile(message)

		// If time.Now() is before "delay-until", keep delaying
		if val, ok := message.Message.Headers["delay-until"]; ok {
			if val2, ok2 := val.(int64); ok2 {
				if val2 > time.Now().Unix() {
					sendToRetryQueue(message)
					continue
				}
			}
		}

		// If first seen time is before incremental backoff
		var min = time.Second * 10
		var max = time.Hour * 6

		var seconds float64
		seconds = math.Pow(2, float64(message.Attempt()))
		seconds = math.Min(seconds, max.Seconds())
		seconds = math.Max(seconds, min.Seconds())

		// Requeue
		if message.FirstSeen().Add(time.Second * time.Duration(int64(seconds))).Before(time.Now()) {

			if seconds >= max.Seconds() {
				log.Warning("Max delay: " + string(message.LastQueue()) + ": " + string(message.Message.Body))
			}

			sendToLastQueue(message)
		} else {
			sendToRetryQueue(message)
		}
	}
}

//noinspection GoUnusedFunction
func writeToFile(message *rabbit.Message) {

	queue := message.Message.Headers["first-queue"]

	f, err := os.OpenFile("delay.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer func() {
		err := f.Close()
		log.Err(err)
	}()

	if queue == nil {
		queue = "None"
	}

	if _, err = f.WriteString(queue.(string) + " - " + string(message.Message.Body) + "\n"); err != nil {
		panic(err)
	}
}
