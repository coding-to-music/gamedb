package consumers

import (
	"math"
	"time"

	"github.com/gamedb/gamedb/pkg/consumers/framework"
)

func delayHandler(messages []framework.Message) {

	for _, message := range messages {

		time.Sleep(time.Second / 10)

		// Get the last seen header
		lastSeenTime := time.Time{}

		if lastSeenVal, ok := message.Message.Headers[framework.HeaderLastSeen]; ok {

			if lastSeenInt, ok2 := lastSeenVal.(int64); ok2 {

				lastSeenTime = time.Unix(lastSeenInt, 0)
			}
		}

		// Get attempt header
		lastSeenReal := 1

		if attemptVal, ok := message.Message.Headers[framework.HeaderAttempt]; ok {

			if lastSeenInt, ok2 := attemptVal.(int); ok2 {

				lastSeenReal = lastSeenInt
			}
		}

		// Get retry time
		var min = time.Second * 2
		var max = time.Hour

		var seconds float64
		seconds = math.Pow(1.5, float64(lastSeenReal))
		seconds = math.Max(seconds, min.Seconds())
		seconds = math.Min(seconds, max.Seconds())

		retryTime := lastSeenTime.Add(time.Second * time.Duration(int64(seconds)))

		// Requeue
		if retryTime.Before(time.Now()) {
			sendToFirstQueue(message)
		} else {
			sendToBackOfQueue(message)
		}
	}
}
