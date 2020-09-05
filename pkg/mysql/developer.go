package mysql

import (
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
)

type Developer struct {
	ID        int        `gorm:"not null;primary_key;AUTO_INCREMENT"`
	CreatedAt time.Time  `gorm:"not null"`
	UpdatedAt time.Time  `gorm:"not null"`
	DeletedAt *time.Time `gorm:""`
	Name      string     `gorm:"not null;index:name"`
	Apps      int        `gorm:"not null"`
	MeanPrice string     `gorm:"not null"` // map[steamapi.CountryCode]float64
	MeanScore float64    `gorm:"not null"`
}

func (d Developer) GetPath() string {
	return "/games?developers=" + strconv.Itoa(d.ID)
}

func (d Developer) GetName() (name string) {
	if d.Name == "" {
		return "Developer " + humanize.Comma(int64(d.ID))
	}

	return d.Name
}

func (d Developer) GetMeanPrice(code steamapi.ProductCC) (string, error) {
	return GetMeanPrice(code, d.MeanPrice)
}

func (d Developer) GetMeanScore() string {
	return helpers.FloatToString(d.MeanScore, 2) + "%"
}

func GetDevelopersByID(ids []int, columns []string) (developers []Developer, err error) {

	if len(ids) == 0 {
		return developers, err
	}

	db, err := GetMySQLClient()
	if err != nil {
		return developers, err
	}

	if len(columns) > 0 {
		db = db.Select(columns)
	}

	db = db.Where("id IN (?)", ids)
	db = db.Order("name ASC")
	db = db.Limit(100)
	db = db.Find(&developers)

	return developers, db.Error
}

func GetAllDevelopers(fields []string) (developers []Developer, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return developers, err
	}

	if len(fields) > 0 {
		db = db.Select(fields)
	}

	db = db.Find(&developers)
	if db.Error != nil {
		return developers, db.Error
	}

	return developers, nil
}
