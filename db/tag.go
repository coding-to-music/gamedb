package db

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/memcache"
)

type Tag struct {
	ID           int        `gorm:"not null;primary_key;AUTO_INCREMENT"`
	CreatedAt    *time.Time `gorm:"not null"`
	UpdatedAt    *time.Time `gorm:"not null"`
	DeletedAt    *time.Time `gorm:""`
	Name         string     `gorm:"not null;index:name"`
	Apps         int        `gorm:"not null"`
	MeanPrice    float64    `gorm:"not null"`
	MeanDiscount float64    `gorm:"not null"`
}

func (tag Tag) GetPath() string {
	return "/games?tag=" + strconv.Itoa(tag.ID)
}

func (tag Tag) GetName() (name string) {

	if tag.Name == "" {
		tag.Name = "Tag " + strconv.Itoa(tag.ID)
	}

	return tag.Name
}

func (tag Tag) GetMeanPrice() float64 {
	return helpers.CentsFloat(tag.MeanPrice)
}

func (tag Tag) GetMeanDiscount() float64 {
	return helpers.DollarsFloat(tag.MeanDiscount)
}

func GetCoopTags() []int {
	return []int{
		1685, // Co-op
		3843, // Online co-op
		3841, // Local co-op
		4508, // Co-op campaign

		3859,  // Multiplayer
		128,   // Massively multiplayer
		7368,  // Local multiplayer
		17770, // Asynchronous multiplayer
	}
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

	err = json.Unmarshal([]byte(s), &tags)
	return tags, err
}

func GetTagsByID(ids []int) (tags []Tag, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return tags, err
	}

	db = db.Limit(100).Where("id IN (?)", ids).Order("name ASC").Find(&tags)
	if db.Error != nil {
		return tags, db.Error
	}

	return tags, nil
}

func SaveOrUpdateTag(id int, vals Tag) (err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	tag := new(Tag)
	db.Assign(vals).FirstOrCreate(tag, Tag{ID: id})
	if db.Error != nil {
		return db.Error
	}

	return nil
}

func DeleteTag(id int) (err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	tag := new(Tag)
	tag.ID = id

	db.Delete(tag)
	if db.Error != nil {
		return db.Error
	}

	return nil
}
