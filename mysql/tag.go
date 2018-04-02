package mysql

import (
	"fmt"
	"strconv"
	"time"
)

type Tag struct {
	ID           int        `gorm:"not null;column:id;primary_key;AUTO_INCREMENT"`
	CreatedAt    *time.Time `gorm:"not null;column:created_at"`
	UpdatedAt    *time.Time `gorm:"not null;column:updated_at"`
	Name         string     `gorm:"not null;column:name"`
	Apps         int        `gorm:"not null;column:apps"`
	MeanPrice    float64    `gorm:"not null;column:mean_price"`
	MeanDiscount float64    `gorm:"not null;column:mean_discount"`
}

func (tag Tag) GetPath() string {
	return "/apps?tag=" + strconv.Itoa(tag.ID)
}

func (tag Tag) GetName() (name string) {

	if tag.Name == "" {
		tag.Name = "Tag " + strconv.Itoa(tag.ID)
	}

	return tag.Name
}

func (tag Tag) GetMeanPrice() string {
	return fmt.Sprintf("%0.2f", tag.MeanPrice/100)
}

func (tag Tag) GetMeanDiscount() string {
	return fmt.Sprintf("%0.2f", tag.MeanDiscount)
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

	db, err := GetDB()
	if err != nil {
		return tags, err
	}

	db = db.Limit(1000).Order("name ASC").Find(&tags)
	if db.Error != nil {
		return tags, db.Error
	}

	return tags, nil
}

func GetTagsByID(ids []int) (tags []Tag, err error) {

	db, err := GetDB()
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

	db, err := GetDB()
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
