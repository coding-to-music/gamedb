package generated

import (
	"math"
)

func (pagination *PaginationSchema) Fill(offset, limit, count int64) {
	pagination.Offset = offset
	pagination.Limit = limit
	pagination.Total = count
	pagination.PagesTotal = int(math.Ceil(float64(count) / float64(limit)))
	pagination.PagesCurrent = int(math.Ceil(float64(offset)/float64(limit))) + 1
}
