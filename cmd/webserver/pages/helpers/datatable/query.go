package datatable

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/derekstavis/go-qs"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/jinzhu/gorm"
	"go.mongodb.org/mongo-driver/bson"
)

// DataTablesQuery
func NewDataTableQuery(r *http.Request, limit bool) (query DataTablesQuery) {

	// Convert string into map
	queryMap, err := qs.Unmarshal(r.URL.Query().Encode())
	if err != nil {
		log.Err(err)
		return
	}

	// Convert map into struct
	err = helpers.MarshalUnmarshal(queryMap, &query)
	if err != nil {
		log.Err(err)
		return
	}

	if limit {

		query.limited = true

		maxStart := sql.UserLevel(session.GetUserLevel(r)).MaxOffset(100)
		start, _ := strconv.Atoi(query.Start)

		if maxStart > 0 && int64(start) > maxStart {

			query.Start = strconv.FormatInt(int64(start), 10)
		}
	}

	return query
}

type DataTablesQuery struct {
	Draw   string                            `json:"draw"`
	Order  map[string]map[string]interface{} `json:"order"`
	Start  string                            `json:"start"`
	Search map[string]interface{}            `json:"search"`
	// Time   string `json:"_"`
	// Columns []string
	limited bool
}

func (q DataTablesQuery) GetSearchString(k string) (search string) {

	if val, ok := q.Search[k]; ok {
		if ok && val != "" {
			if val, ok := val.(string); ok {
				if ok {
					return val
				}
			}
		}
	}

	return ""
}

func (q DataTablesQuery) GetSearchSlice(k string) (search []string) {

	if val, ok := q.Search[k]; ok {
		if val != "" {

			if val, ok := val.([]interface{}); ok {
				for k, v := range val {
					if val2, ok2 := v.(string); ok2 {
						if k < 10 { // Limit to 10 items
							search = append(search, val2)
						}
					}
				}
			}
		}
	}

	return search
}

func (q DataTablesQuery) GetOrderSQL(columns map[string]string) (order string) {

	var orders []string
	for _, v := range q.getOrder(columns) {

		var sort string
		if v.asc {
			sort = "asc"
		} else {
			sort = "desc"
		}

		orders = append(orders, v.col+" "+sort)
	}

	return strings.Join(orders, ", ")
}

func (q DataTablesQuery) GetOrderMongo(columns map[string]string) (cols bson.D) {

	for _, v := range q.getOrder(columns) {

		var sort int
		if v.asc {
			sort = 1
		} else {
			sort = -1
		}

		cols = append(cols, bson.E{Key: v.col, Value: sort})
	}

	return cols
}

func (q DataTablesQuery) GetOrderElastic(columns map[string]string) (string, bool) {

	for _, v := range q.getOrder(columns) {
		return v.col, v.asc
	}

	return "_score", false
}

func (q DataTablesQuery) getOrder(colsMap map[string]string) (colsRet []sortCol) {

	for _, v := range q.Order {

		if col, ok := v["column"].(string); ok {
			if ok {

				if dir, ok := v["dir"].(string); ok {
					if ok {

						if cols, ok := colsMap[col]; ok {
							if ok {

								for k, col := range strings.Split(cols, ",") {

									col = strings.TrimSpace(col)
									colParts := strings.Split(col, " ")

									if k == 0 {
										colsRet = append(colsRet, sortCol{col: col, asc: dir != "desc"})
									} else if len(colParts) == 1 {
										colsRet = append(colsRet, sortCol{col: col, asc: true})
									} else if len(colParts) == 2 {
										colsRet = append(colsRet, sortCol{col: colParts[0], asc: colParts[1] != "desc"})
									} else {
										log.Warning("weird column map")
									}

									continue
								}

								// Only get first sort from url
								return
							}
						}
					}
				}
			}
		}
	}

	return colsRet
}

func (q DataTablesQuery) SetOrderOffsetGorm(db *gorm.DB, columns map[string]string) *gorm.DB {

	db = db.Order(q.GetOrderSQL(columns))
	db = db.Offset(q.Start)

	return db
}

func (q DataTablesQuery) GetOffset() int {
	i, _ := strconv.Atoi(q.Start)
	return i
}

func (q DataTablesQuery) GetOffset64() int64 {
	i, _ := strconv.ParseInt(q.Start, 10, 64)
	return i
}

func (q DataTablesQuery) GetPage(perPage int) int {

	i, _ := strconv.Atoi(q.Start)

	if i == 0 {
		return 1
	}

	return (i / perPage) + 1
}

type sortCol struct {
	col string
	asc bool
}
