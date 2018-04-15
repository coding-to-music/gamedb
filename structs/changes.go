package structs

import (
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/mysql"
)

type ChangesChangeTemplate struct {
	Change   datastore.Change
	Apps     []mysql.App
	Packages []mysql.Package
}

type ChangesChangeAppTemplate struct {
	ID   int
	Name string
}
