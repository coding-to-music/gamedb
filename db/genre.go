package db

import (
	"encoding/json"
	"time"

	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/memcache"
)

type Genre struct {
	ID        int        `gorm:"not null;primary_key;AUTO_INCREMENT"`
	CreatedAt *time.Time `gorm:"not null"`
	UpdatedAt *time.Time `gorm:"not null"`
	DeletedAt *time.Time `gorm:""`
	Name      string     `gorm:"not null;index:name"`
	Apps      int        `gorm:"not null"`
	MeanPrice float64    `gorm:"not null"`
	MeanScore float64    `gorm:"not null"`
}

func (g Genre) GetPath() string {
	return "/games?genre=" + g.Name
}

func (g Genre) GetName() string {
	return g.Name
}

func (g Genre) GetMeanPrice() float64 {
	return helpers.CentsFloat(g.MeanPrice)
}

func (g Genre) GetMeanScore() float64 {
	return helpers.DollarsFloat(g.MeanScore)
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

	s, err := memcache.GetSetString(memcache.GenreKeyNames, func() (s string, err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return s, err
		}

		var genres []Genre
		db = db.Select([]string{"id", "name"}).Order("name ASC").Find(&genres)
		if db.Error != nil {
			return s, db.Error
		}

		bytes, err := json.Marshal(genres)
		return string(bytes), err
	})

	if err != nil {
		return genres, err
	}

	err = json.Unmarshal([]byte(s), &genres)
	return genres, err
}

func SaveOrUpdateGenre(id int, name string, apps int) (err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	genre := new(Genre)
	db.Attrs(Genre{Name: name}).Assign(Genre{Apps: apps}).FirstOrCreate(genre, Genre{ID: id})
	if db.Error != nil {
		return db.Error
	}

	return nil
}

func DeleteGenre(id int) (err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	genre := new(Genre)
	genre.ID = id

	db.Delete(genre)
	if db.Error != nil {
		return db.Error
	}

	return nil
}
