package db

import (
	"time"

	"github.com/steam-authority/steam-authority/helpers"
)

type Genre struct {
	ID           int        `gorm:"not null;primary_key;AUTO_INCREMENT"` //
	CreatedAt    *time.Time `gorm:"not null"`
	UpdatedAt    *time.Time `gorm:"not null"`
	DeletedAt    *time.Time `gorm:""`
	Name         string     `gorm:"not null;index:name"`
	Apps         int        `gorm:"not null"`
	MeanPrice    float64    `gorm:"not null"`
	MeanDiscount float64    `gorm:"not null"`
}

func (g Genre) GetPath() string {
	return "/games?genre=" + g.Name
}

func (g Genre) GetMeanPrice() float64 {
	return helpers.CentsFloat(g.MeanPrice)
}

func (g Genre) GetMeanDiscount() float64 {
	return helpers.DollarsFloat(g.MeanDiscount)
}

func GetAllGenres() (genres []Genre, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return genres, err
	}

	db.Limit(1000).Find(&genres)
	if db.Error != nil {
		return genres, db.Error
	}

	return genres, nil
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
