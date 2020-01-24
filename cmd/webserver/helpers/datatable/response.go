package datatable

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/sql"
)

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

func (t *DataTablesResponse) Output() [][]interface{} {

	if len(t.Data) == 0 {
		t.Data = make([][]interface{}, 0)
	}
	return t.Data
}

func (t *DataTablesResponse) Limit(r *http.Request) {

	level := sql.UserLevel(helpers.GetUserLevel(r))
	max := level.MaxResults(100)

	if max > 0 && max < t.RecordsFiltered {
		t.RecordsFiltered = max
		t.LevelLimited = true
	}
}

