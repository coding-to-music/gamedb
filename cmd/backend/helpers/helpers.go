package helpers

import (
	"math"

	"github.com/gamedb/gamedb/pkg/backend/generated"
	"go.mongodb.org/mongo-driver/bson"
)

func MakePaginationResponse(req *generated.PaginationRequest, total int64, totalFiltered int64) *generated.PaginationResponse {

	m := &generated.PaginationResponse{}
	m.Offset = req.GetOffset()
	m.Limit = req.GetLimit()
	m.Total = total
	m.TotalFiltered = totalFiltered
	m.PagesTotal = int64(math.Ceil(float64(total) / float64(req.GetLimit())))
	m.PagesCurrent = int64(math.Floor(float64(req.GetOffset())/float64(req.GetLimit())) + 1)

	return m
}

func MakeMongoOrder(request *generated.PaginationRequest) (o bson.D) {

	field := request.GetSortField()

	if field == "" {
		return bson.D{}
	}

	order := 1
	if request.GetSortOrder() == "desc" {
		order = -1
	}

	return bson.D{{field, order}}
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
