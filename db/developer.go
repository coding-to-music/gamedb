package db

import (
	"encoding/json"
	"sort"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/memcache"
)

type Developer struct {
	ID        int        `gorm:"not null;primary_key;AUTO_INCREMENT"`
	CreatedAt *time.Time `gorm:"not null"`
	UpdatedAt *time.Time `gorm:"not null"`
	DeletedAt *time.Time `gorm:""`
	Name      string     `gorm:"not null;index:name"`
	Apps      int        `gorm:"not null"`
	MeanPrice string     `gorm:"not null"`
	MeanScore string     `gorm:"not null"`
}

func (d Developer) GetPath() string {
	return "/games?developers=" + strconv.Itoa(d.ID)
}

func (d Developer) GetName() (name string) {
	return d.Name
}

func (d Developer) GetMeanPrice(code steam.CountryCode) (string, error) {
	return helpers.GetMeanPrice(code, d.MeanPrice)
}

func (d Developer) GetMeanScore(code steam.CountryCode) (string, error) {
	return helpers.GetMeanScore(code, d.MeanScore)
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
		db = db.Select([]string{"id", "name"}).Order("apps DESC").Limit(500).Find(&devs)
		if db.Error != nil {
			return s, db.Error
		}

		sort.Slice(devs, func(i, j int) bool {
			return devs[i].Name < devs[j].Name
		})

		bytes, err := json.Marshal(devs)
		return string(bytes), err
	})

	if err != nil {
		return devs, err
	}

	err = helpers.Unmarshal([]byte(s), &devs)
	return devs, err
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
