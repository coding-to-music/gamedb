package generated

import (
	"math"
)

func (m *PaginationResponse) SetPagination(req *PaginationRequest, total int64) {

	m.Offset = req.GetOffset()
	m.Limit = req.GetLimit()
	m.Total = total
	m.PagesTotal = int64(math.Ceil(float64(total) / float64(req.GetLimit())))
	m.PagesCurrent = int64(math.Floor(float64(req.GetOffset())/float64(req.GetLimit())) + 1)
}
