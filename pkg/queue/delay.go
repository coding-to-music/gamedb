package queue

import (
	"math"
	"time"

	"github.com/Jleagle/rabbit-go"
)

func delayHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		time.Sleep(time.Second / 10)

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
			sendToLastQueue(message)
		} else {
			sendToRetryQueue(message)
		}
	}
}
