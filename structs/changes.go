package structs

import (
	"github.com/steam-authority/steam-authority/db"
)

type ChangesChangeTemplate struct {
	Change   db.Change
	Apps     []db.App
	Packages []db.Package
}

type ChangesChangeAppTemplate struct {
	ID   int
	Name string
}
