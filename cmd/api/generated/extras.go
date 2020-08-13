package generated

import (
	"math"

	"github.com/gamedb/gamedb/pkg/backend"
)

func (pagination *PaginationSchema) Fill(offset, limit, count int64) {
	pagination.Offset = offset
	pagination.Limit = limit
	pagination.Total = count
	pagination.PagesTotal = int64(math.Ceil(float64(count) / float64(limit)))
	pagination.PagesCurrent = int64(math.Floor(float64(offset)/float64(limit)) + 1)
}

func (pagination *PaginationSchema) FillFromProto(m backend.PaginationResponse) {
	pagination.Offset = m.GetOffset()
	pagination.Limit = m.GetLimit()
	pagination.Total = m.GetTotal()
	pagination.PagesTotal = m.GetPagesTotal()
	pagination.PagesCurrent = m.GetPagesCurrent()
}
