package helpers

import (
	"strconv"

	"github.com/gosimple/slug"
)

func IsValidPackageID(id int) bool {
	return id != 0
}

func GetPackagePath(id int, name string) string {

	path := "/packages/" + strconv.Itoa(id)
	if name == "" {
		return path
	}
	return path + "/" + slug.Make(name)
}
