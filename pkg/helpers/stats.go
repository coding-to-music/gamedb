package helpers

import (
	"strconv"

	"github.com/gosimple/slug"
)

func GetStatPath(p string, id int, name string) string {

	return "/" + p + "/" + strconv.Itoa(id) + "/" + slug.Make(name)
}
