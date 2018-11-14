package db

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/memcache"
)

type Tag struct {
	ID        int        `gorm:"not null;primary_key"`
	CreatedAt *time.Time `gorm:"not null"`
	UpdatedAt *time.Time `gorm:"not null"`
	DeletedAt *time.Time `gorm:""`
	Name      string     `gorm:"not null;index:name"`
	Apps      int        `gorm:"not null"`
	MeanPrice string     `gorm:"not null"`
	MeanScore string     `gorm:"not null"`
}

func (tag Tag) GetPath() string {
	return "/games?tags=" + strconv.Itoa(tag.ID)
}

func (tag Tag) GetName() (name string) {

	if tag.Name == "" {
		return "Tag " + strconv.Itoa(tag.ID)
	}

	return tag.Name
}

func (tag Tag) GetMeanPrice(code steam.CountryCode) (string, error) {
	return helpers.GetMeanPrice(code, tag.MeanPrice)
}

func (tag Tag) GetMeanScore(code steam.CountryCode) (string, error) {
	return helpers.GetMeanScore(code, tag.MeanScore)
}

func GetAllTags() (tags []Tag, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return tags, err
	}

	db = db.Find(&tags)
	if db.Error != nil {
		return tags, db.Error
	}

	return tags, nil
}

func GetTagsForSelect() (tags []Tag, err error) {

	s, err := memcache.GetSetString(memcache.TagKeyNames, func() (s string, err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return s, err
		}

		var tags []Tag
		db = db.Select([]string{"id", "name"}).Order("name ASC").Find(&tags)
		if db.Error != nil {
			return s, db.Error
		}

		bytes, err := json.Marshal(tags)
		return string(bytes), err
	})

	if err != nil {
		return tags, err
	}

	err = helpers.Unmarshal([]byte(s), &tags)
	return tags, err
}

func GetTagsByID(ids []int) (tags []Tag, err error) {

	if len(ids) == 0 {
		return tags, err
	}

	db, err := GetMySQLClient()
	if err != nil {
		return tags, err
	}

	db = db.Limit(100).Where("id IN (?)", ids).Order("name ASC").Find(&tags)

	return tags, db.Error
}

func DeleteTags(ids []int) (err error) {

	fmt.Println("Deleteing " + strconv.Itoa(len(ids)) + " tags")

	if len(ids) == 0 {
		return nil
	}

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	db.Where("id IN (?)", ids).Delete(Tag{})

	return db.Error
}
