package helpers

import (
	"encoding/json"
	"time"

	"github.com/Jleagle/go-durationfmt"
	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
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

		var is = []zap.Field{
			zap.String("method", ql.method),
			zap.String("collection", ql.collection),
		}

		b, _ := json.Marshal(ql.filter)
		is = append(is, zap.ByteString("filter", b))

		b, _ = json.Marshal(ql.sort)
		is = append(is, zap.ByteString("sort", b))

		zap.L().Named(log.LogNameMongo).Warn("Mongo call taking "+diffFormatted, is...)
	}
}
