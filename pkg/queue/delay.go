package queue

import (
	"math"
	"time"

	"github.com/Jleagle/rabbit-go"
)

func delayHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		time.Sleep(time.Second / 10)

		var min = time.Second * 2
		var max = time.Hour

		var seconds float64
		seconds = math.Pow(1.5, float64(message.Attempt()))
		seconds = math.Max(seconds, min.Seconds())
		seconds = math.Min(seconds, max.Seconds())

		// Requeue
		if message.FirstSeen().Add(time.Second * time.Duration(int64(seconds))).Before(time.Now()) {
			sendToRetryQueue(message)
		} else {
			sendToLastQueue(message)
		}
	}
}
