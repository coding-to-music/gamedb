package helpers

import (
	"encoding/json"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
)

type QueryLogger struct {
	startTime time.Time
	filter    interface{}
	sort      interface{}
}

func (ql QueryLogger) Start(filter interface{}, sort interface{}) {
	ql.startTime = time.Now()
	ql.filter = filter
	ql.sort = sort
}

func (ql QueryLogger) End() {

	diff := time.Now().Sub(ql.startTime)

	if diff > (time.Second * 1) {

		var is = []interface{}{
			log.LogNameMongo,
			"Mongo call taking " + diff.String(),
		}

		b, _ := json.Marshal(ql.filter)
		is = append(is, "Filter: "+string(b))

		b, _ = json.Marshal(ql.sort)
		is = append(is, "Sort: "+string(b))

		log.Info(is...)
	}
}
