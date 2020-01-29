package datatable

import (
	"go.mongodb.org/mongo-driver/bson"
)

type Columns map[string]Column

type Column struct {
	sortAsc     bool
	sortDesc    bool
	sortDefault bson.D
	sortAppend  bson.D
	filters     bson.D
}
