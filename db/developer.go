package db

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/memcache"
)

type Developer struct {
	ID        int        `gorm:"not null;primary_key;AUTO_INCREMENT"`
	CreatedAt *time.Time `gorm:"not null"`
	UpdatedAt *time.Time `gorm:"not null"`
	DeletedAt *time.Time `gorm:""`
	Name      string     `gorm:"not null;index:name"`
	Apps      int        `gorm:"not null"`
	MeanPrice float64    `gorm:"not null"`
	MeanScore float64    `gorm:"not null"`
}

func (d Developer) GetPath() string {
	return "/games?developer=" + strconv.Itoa(d.ID)
}

func (d Developer) GetName() (name string) {
	return d.Name
}

func (d Developer) GetMeanPrice() float64 {
	return helpers.CentsFloat(d.MeanPrice)
}

func (d Developer) GetMeanScore() float64 {
	return helpers.DollarsFloat(d.MeanScore)
}

func GetAllDevelopers() (developers []Developer, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return developers, err
	}

	db = db.Find(&developers)
	if db.Error != nil {
		return developers, db.Error
	}

	return developers, nil
}

func GetDevelopersForSelect() (devs []Developer, err error) {

	s, err := memcache.GetSetString(memcache.DeveloperKeyNames, func() (s string, err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return s, err
		}

		var devs []Developer
		db = db.Select([]string{"id", "name"}).Order("name ASC").Find(&devs)
		if db.Error != nil {
			return s, db.Error
		}

		bytes, err := json.Marshal(devs)
		return string(bytes), err
	})

	if err != nil {
		return devs, err
	}

	err = json.Unmarshal([]byte(s), &devs)
	return devs, err
}

func SaveOrUpdateDeveloper(name string, vals Developer) (err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	developer := new(Developer)
	developer.DeletedAt = nil

	db.Assign(vals).FirstOrCreate(developer, Developer{Name: name})
	if db.Error != nil {
		return db.Error
	}

	return nil
}

func DeleteDeveloper(id int) (err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	developer := new(Developer)
	developer.ID = id

	db.Delete(developer)
	if db.Error != nil {
		return db.Error
	}

	return nil
}

func DeleteDevelopers(ids []int) (err error) {

	if len(ids) == 0 {
		return nil
	}

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	db.Where("id IN (?)", ids).Delete(Developer{})

	return db.Error
}
