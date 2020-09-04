package helpers

import (
	"strconv"

	"github.com/gosimple/slug"
)

func GetStatPath(typex string, id int32, name string) string {

	return "/tags/" + strconv.Itoa(int(id)) + "/" + slug.Make(name)
}
