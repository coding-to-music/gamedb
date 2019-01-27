package db

import (
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
)

type Genre struct {
	ID        int        `gorm:"not null;primary_key;AUTO_INCREMENT"`
	CreatedAt time.Time  `gorm:"not null"`
	UpdatedAt time.Time  `gorm:"not null"`
	DeletedAt *time.Time `gorm:""`
	Name      string     `gorm:"not null;index:name"`
	Apps      int        `gorm:"not null"`
	MeanPrice string     `gorm:"not null"` // JSON
	MeanScore float64    `gorm:"not null"`
}

func (g Genre) GetPath() string {
	return "/apps?genres=" + strconv.Itoa(g.ID)
}

func (g Genre) GetName() string {
	return g.Name
}

func (g Genre) GetMeanPrice(code steam.CountryCode) (string, error) {
	return helpers.GetMeanPrice(code, g.MeanPrice)
}

func (g Genre) GetMeanScore() string {
	return helpers.FloatToString(g.MeanScore, 2) + "%"
}

func GetAllGenres() (genres []Genre, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return genres, err
	}

	db.Find(&genres)
	if db.Error != nil {
		return genres, db.Error
	}

	return genres, nil
}

func GetGenresForSelect() (genres []Genre, err error) {

	var item = helpers.MemcacheGenreKeyNames

	err = helpers.GetMemcache().GetSet(item.Key, item.Expiration, &genres, func() (s interface{}, err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return s, err
		}

		var genres []Genre
		db = db.Select([]string{"id", "name"}).Order("name ASC").Find(&genres)
		return genres, db.Error
	})

	return genres, err
}

func DeleteGenres(ids []int) (err error) {

	if len(ids) == 0 {
		return nil
	}

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	db.Where("id IN (?)", ids).Delete(Genre{})

	return db.Error
}
