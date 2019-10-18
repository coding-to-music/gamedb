package helpers

import (
	"encoding/json"
	"time"

	"github.com/Jleagle/go-durationfmt"
	"github.com/gamedb/gamedb/pkg/log"
)

type QueryLogger struct {
	startTime  time.Time
	collection string
	filter     interface{}
	sort       interface{}
}

func (ql QueryLogger) Start(collection string, filter interface{}, sort interface{}) {
	ql.startTime = time.Now()
	ql.collection = collection
	ql.filter = filter
	ql.sort = sort
}

func (ql QueryLogger) End() {

	diff := time.Now().Sub(ql.startTime)

	if diff > (time.Second * 1) {

		diffFormatted, err := durationfmt.Format(diff, "%mmin %ssec %imil")
		if err != nil {
			diffFormatted = diff.String()
		}

		var is = []interface{}{
			log.LogNameMongo,
			"Mongo call taking " + diffFormatted,
			ql.collection,
		}

		b, _ := json.Marshal(ql.filter)
		is = append(is, "Filter: "+string(b))

		b, _ = json.Marshal(ql.sort)
		is = append(is, "Sort: "+string(b))

		log.Warning(is...)
	}
}
