package helpers

import (
	"math"

	"github.com/gamedb/gamedb/pkg/backend/generated"
)

func MakePagination(req *generated.PaginationRequest, total int64) *generated.PaginationResponse {

	m := &generated.PaginationResponse{}
	m.Offset = req.GetOffset()
	m.Limit = req.GetLimit()
	m.Total = total
	m.PagesTotal = int64(math.Ceil(float64(total) / float64(req.GetLimit())))
	m.PagesCurrent = int64(math.Floor(float64(req.GetOffset())/float64(req.GetLimit())) + 1)

	return m
}

func StringsToInterfaces(s []string) (o []interface{}) {
	for _, v := range s {
		o = append(o, v)
	}
	return o
}

func IntsToInt32s(s []int) (o []int32) {
	for _, v := range s {
		o = append(o, int32(v))
	}
	return o
}
