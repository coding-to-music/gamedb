package helpers

import (
	"encoding/json"
	"time"

	"github.com/Jleagle/go-durationfmt"
	"github.com/gamedb/gamedb/pkg/log"
)

type QueryLogger struct {
	startTime  time.Time
	method     string
	collection string
	filter     interface{}
	sort       interface{}
}

func (ql *QueryLogger) Start(method string, collection string, filter interface{}, sort interface{}) {
	ql.startTime = time.Now()
	ql.method = method
	ql.collection = collection
	ql.filter = filter
	ql.sort = sort
}

func (ql QueryLogger) End() {

	diff := time.Since(ql.startTime)

	if diff > (time.Second * 5) {

		diffFormatted, err := durationfmt.Format(diff, "%s.%is")
		if err != nil {
			diffFormatted = diff.String()
		}

		var is = []interface{}{
			log.LogNameMongo,
			"Mongo call taking " + diffFormatted,
			ql.method,
			ql.collection,
		}

		b, _ := json.Marshal(ql.filter)
		is = append(is, string(b))

		b, _ = json.Marshal(ql.sort)
		is = append(is, string(b))

		log.Warning(is...)
	}
}
