package queue

import (
	"math"
	"time"

	"github.com/Jleagle/rabbit-go"
)

func delayHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		time.Sleep(time.Second / 10)

		// Delay
		if val, ok := message.Message.Headers["delay-until"]; ok {
			if val2, ok2 := val.(int64); ok2 {
				if val2 > time.Now().Unix() {
					sendToRetryQueue(message)
					continue
				}
			}
		}

		var seconds float64
		var max = time.Hour * 6

		seconds = math.Pow(2.5, float64(message.Attempt()))
		seconds = math.Min(seconds, max.Seconds())

		// Requeue
		if message.FirstSeen().Add(time.Second * time.Duration(int64(seconds))).Before(time.Now()) {
			sendToLastQueue(message)
		} else {
			sendToRetryQueue(message)
		}
	}
}
