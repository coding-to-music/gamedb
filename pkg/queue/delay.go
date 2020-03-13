package queue

import (
	"math"
	"time"

	"github.com/Jleagle/rabbit-go"
)

func delayHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		time.Sleep(time.Second / 10)

		var max = time.Hour * 6

		var seconds float64
		seconds = math.Pow(2.5, float64(message.Attempt()))
		seconds = math.Min(seconds, max.Seconds())

		// Requeue
		if message.LastSeen().Add(time.Second * time.Duration(int64(seconds))).Before(time.Now()) {
			sendToRetryQueue(message)
		} else {
			sendToLastQueue(message)
		}
	}
}
