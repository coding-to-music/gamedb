package sql

import (
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
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
	return "/apps?genres=" + strconv.Itoa(g.ID)
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

func GetGenresForSelect() (genres []Genre, err error) {

	var item = memcache.MemcacheGenreKeyNames

	err = memcache.GetSetInterface(item.Key, item.Expiration, &genres, func() (interface{}, error) {

		var genres []Genre

		db, err := GetMySQLClient()
		if err != nil {
			return genres, err
		}

		db = db.Select([]string{"id", "name"}).Order("name ASC").Find(&genres)
		return genres, db.Error
	})

	return genres, err
}

func GetGenresByID(ids []int, columns []string) (genres []Genre, err error) {

	if len(ids) == 0 {
		return genres, err
	}

	db, err := GetMySQLClient()
	if err != nil {
		return genres, err
	}

	if len(columns) > 0 {
		db = db.Select(columns)
	}

	db = db.Where("id IN (?)", ids)
	db = db.Order("name ASC")
	db = db.Limit(100)
	db = db.Find(&genres)

	return genres, db.Error
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
