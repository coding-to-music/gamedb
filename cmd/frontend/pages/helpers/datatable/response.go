package datatable

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/frontend/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/mysql"
)

func NewDataTablesResponse(r *http.Request, query DataTablesQuery, count int64, countFiltered int64, aggregations map[string]map[string]int64) (ret *DataTablesResponse) {

	if count < countFiltered {
		count = countFiltered
	}

	ret = &DataTablesResponse{}
	ret.Draw = query.Draw
	ret.Data = make([][]interface{}, 0)
	ret.RecordsTotal = count
	ret.RecordsFiltered = countFiltered
	ret.Aggregations = aggregations

	if query.limited {

		level := mysql.UserLevel(session.GetUserLevel(r))
		max := level.MaxResults(100)

		if max > 0 && max < ret.RecordsFiltered {

			ret.RecordsFiltered = max

			if session.IsLoggedIn(r) {
				ret.LevelLimited = 1
			} else {
				ret.LevelLimited = 2
			}
		}
	}

	return ret
}

// DataTablesResponse
type DataTablesResponse struct {
	Draw            string                      `json:"draw"`
	RecordsTotal    int64                       `json:"recordsTotal,string"`
	RecordsFiltered int64                       `json:"recordsFiltered,string"`
	LevelLimited    int                         `json:"limited"` // 0 - Not limited, 1 - logged in, 2 - guest
	Data            [][]interface{}             `json:"data"`
	Aggregations    map[string]map[string]int64 `json:"aggregations"`
}

func (t *DataTablesResponse) AddRow(row []interface{}) {
	t.Data = append(t.Data, row)
}
