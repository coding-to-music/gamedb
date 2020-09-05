package mysql

import (
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
)

type Genre struct {
	ID        int        `gorm:"not null;primary_key;AUTO_INCREMENT"`
	CreatedAt time.Time  `gorm:"not null"`
	UpdatedAt time.Time  `gorm:"not null"`
	DeletedAt *time.Time `gorm:""`
	Name      string     `gorm:"not null;index:name"`
	Apps      int        `gorm:"not null"`
	MeanPrice string     `gorm:"not null"` // map[steamapi.CountryCode]float64
	MeanScore float64    `gorm:"not null"`
}

func (g Genre) GetPath() string {
	return "/games?genres=" + strconv.Itoa(g.ID)
}

func (g Genre) GetName() string {
	if g.Name == "" {
		return "Genre " + humanize.Comma(int64(g.ID))
	}

	return g.Name
}

func (g Genre) GetMeanPrice(code steamapi.ProductCC) (string, error) {
	return GetMeanPrice(code, g.MeanPrice)
}

func (g Genre) GetMeanScore() string {
	return helpers.FloatToString(g.MeanScore, 2) + "%"
}

func GetAllGenres(includeDeleted bool) (genres []Genre, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return genres, err
	}

	if includeDeleted {
		db = db.Unscoped()
	}

	db = db.Find(&genres)
	if db.Error != nil {
		return genres, db.Error
	}

	return genres, nil
}
