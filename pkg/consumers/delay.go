package consumers

import (
	"math"
	"time"

	"github.com/gamedb/gamedb/pkg/consumers/framework"
)

func delayHandler(messages []*framework.Message) {

	for _, message := range messages {

		time.Sleep(time.Second / 10)

		var min = time.Second * 2
		var max = time.Hour

		var seconds float64
		seconds = math.Pow(1.5, float64(message.Attempt()))
		seconds = math.Max(seconds, min.Seconds())
		seconds = math.Min(seconds, max.Seconds())

		// Requeue
		if message.LastSeen().Add(time.Second * time.Duration(int64(seconds))).Before(time.Now()) {
			sendToFirstQueue(message)
		} else {
			sendToRetryQueue(message)
		}
	}
}
