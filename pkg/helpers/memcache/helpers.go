package memcache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

var ErrInQueue = errors.New("already in queue")

//
func ProjectionToString(m bson.M) string {

	if len(m) == 0 {
		return "*"
	}

	var cols []string
	for k := range m {
		cols = append(cols, k)
	}

	sort.Slice(cols, func(i, j int) bool {
		return cols[i] < cols[j]
	})

	return strings.Join(cols, "-")
}

func FilterToString(d bson.D) string {

	if d == nil || len(d) == 0 {
		return "[]"
	}

	b, err := json.Marshal(d)
	if err != nil {
		log.Err(err)
		return "[]"
	}

	h := md5.Sum(b)

	return hex.EncodeToString(h[:])
}
