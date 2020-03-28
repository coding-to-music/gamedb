package datatable

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/derekstavis/go-qs"
	"github.com/gamedb/gamedb/cmd/webserver/helpers/session"
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

		level := sql.UserLevel(session.GetUserLevel(r))
		max := level.MaxOffset(100)

		start, _ := strconv.Atoi(query.Start)

		if max > 0 && int64(start) > max {
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

func (q DataTablesQuery) GetOrderSQL(columns map[string]string, defaultCol string) (order string) {

	var ret []string

	for _, v := range q.Order {

		col, ok := v["column"].(string)
		if !ok || col == "" {
			col = defaultCol
		}

		if dir, ok := v["dir"].(string); ok {
			if ok {

				if columns != nil {
					col, ok := columns[col]
					if ok {
						if dir == "asc" || dir == "desc" {
							if strings.Contains(col, "$dir") {
								ret = append(ret, strings.Replace(col, "$dir", dir, 1))
							} else {
								ret = append(ret, col+" "+dir)
							}
						}
					}
				}
			}
		}
	}

	return strings.Join(ret, ", ")
}

func (q DataTablesQuery) GetOrderMongo(columns map[string]string) bson.D {

	for _, v := range q.Order {

		if col, ok := v["column"].(string); ok {
			if ok {

				if dir, ok := v["dir"].(string); ok {
					if ok {

						if col, ok := columns[col]; ok {
							if ok {

								if dir == "desc" {
									return bson.D{{Key: col, Value: -1}}
								}

								return bson.D{{Key: col, Value: 1}}
							}
						}
					}
				}
			}
		}
	}

	return bson.D{}
}

func (q DataTablesQuery) GetOrderString(columns map[string]string) (col string) {

	for _, v := range q.Order {

		if col, ok := v["column"].(string); ok {
			if ok {
				if col, ok := columns[col]; ok {
					if ok {
						return col
					}
				}
			}
		}
	}

	return col
}

func (q DataTablesQuery) SetOrderOffsetGorm(db *gorm.DB, columns map[string]string, defaultCol string) *gorm.DB {

	db = db.Order(q.GetOrderSQL(columns, defaultCol))
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
