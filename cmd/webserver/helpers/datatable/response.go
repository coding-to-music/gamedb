package datatable

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/sql"
)

func NewDataTablesResponse(r *http.Request, query DataTablesQuery, count int64, countFiltered int64) (ret *DataTablesResponse) {

	ret = &DataTablesResponse{}
	ret.Draw = query.Draw
	ret.Data = make([][]interface{}, 0)
	ret.RecordsTotal = count
	ret.RecordsFiltered = countFiltered

	if query.limited {

		level := sql.UserLevel(helpers.GetUserLevel(r))
		max := level.MaxResults(100)

		if max > 0 && max < ret.RecordsFiltered {
			ret.RecordsFiltered = max
			ret.LevelLimited = true
		}
	}

	return ret
}

// DataTablesResponse
type DataTablesResponse struct {
	Draw            string          `json:"draw"`
	RecordsTotal    int64           `json:"recordsTotal,string"`
	RecordsFiltered int64           `json:"recordsFiltered,string"`
	LevelLimited    bool            `json:"limited"`
	Data            [][]interface{} `json:"data"`
}

func (t *DataTablesResponse) AddRow(row []interface{}) {
	t.Data = append(t.Data, row)
}
