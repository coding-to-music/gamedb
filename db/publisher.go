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

type Publisher struct {
	ID        int        `gorm:"not null;primary_key;AUTO_INCREMENT"`
	CreatedAt *time.Time `gorm:"not null"`
	UpdatedAt *time.Time `gorm:"not null"`
	DeletedAt *time.Time `gorm:""`
	Name      string     `gorm:"not null;index:name"`
	Apps      int        `gorm:"not null"`
	MeanPrice string     `gorm:"not null"`
	MeanScore string     `gorm:"not null"`
}

func (p Publisher) GetPath() string {
	return "/games?publishers=" + strconv.Itoa(p.ID)
}

func (p Publisher) GetName() (name string) {
	return p.Name
}

func (p Publisher) GetMeanPrice(code steam.CountryCode) (string, error) {
	return helpers.GetMeanPrice(code, p.MeanPrice)
}

func (p Publisher) GetMeanScore(code steam.CountryCode) (string, error) {
	return helpers.GetMeanScore(code, p.MeanScore)
}

func GetPublishersByID(ids []int, columns []string) (publishers []Publisher, err error) {

	if len(ids) == 0 {
		return publishers, nil
	}

	db, err := GetMySQLClient()
	if err != nil {
		return publishers, err
	}

	if len(columns) > 0 {
		db = db.Select(columns)
	}

	db.Where("id IN (?)", ids).Find(&publishers)
	if db.Error != nil {
		return publishers, db.Error
	}

	return publishers, nil
}

func GetAllPublishers() (publishers []Publisher, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return publishers, err
	}

	db = db.Find(&publishers)
	if db.Error != nil {
		return publishers, db.Error
	}

	return publishers, nil
}

func GetPublishersForSelect() (pubs []Publisher, err error) {

	s, err := memcache.GetSetString(memcache.PublisherKeyNames, func() (s string, err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return s, err
		}

		var pubs []Publisher
		db = db.Select([]string{"id", "name"}).Order("apps DESC").Limit(500).Find(&pubs)
		if db.Error != nil {
			return s, db.Error
		}

		sort.Slice(pubs, func(i, j int) bool {
			return pubs[i].Name < pubs[j].Name
		})

		bytes, err := json.Marshal(pubs)
		return string(bytes), err
	})

	if err != nil {
		return pubs, err
	}

	err = helpers.Unmarshal([]byte(s), &pubs)
	return pubs, err
}

func DeletePublishers(ids []int) (err error) {

	if len(ids) == 0 {
		return nil
	}

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	db.Where("id IN (?)", ids).Delete(Publisher{})

	return db.Error
}
