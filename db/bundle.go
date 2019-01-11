package db

import (
	"encoding/json"
	"time"

	"github.com/gamedb/website/helpers"
	"github.com/jinzhu/gorm"
)

type Bundle struct {
	ID         int
	CreatedAt  time.Time `gorm:"not null;column:created_at;type:datetime"`
	UpdatedAt  time.Time `gorm:"not null;column:updated_at;type:datetime"`
	Name       string
	Discount   int
	AppIDs     string
	PackageIDs string
}

func (bundle *Bundle) BeforeCreate(scope *gorm.Scope) error {

	if bundle.AppIDs == "" {
		bundle.AppIDs = "[]"
	}
	if bundle.PackageIDs == "" {
		bundle.PackageIDs = "[]"
	}

	return nil
}

func (bundle Bundle) GetAppIDs() (ids []int, err error) {

	err = helpers.Unmarshal([]byte(bundle.AppIDs), &ids)
	return ids, err
}

func (bundle Bundle) GetPackageIDs() (ids []int, err error) {

	err = helpers.Unmarshal([]byte(bundle.PackageIDs), &ids)
	return ids, err
}

func (bundle *Bundle) SetAppIDs(ids []int) (err error) {

	bytes, err := json.Marshal(ids)
	if err != nil {
		return err
	}

	bundle.AppIDs = string(bytes)
	return nil
}

func (bundle *Bundle) SetPackageIDs(ids []int) (err error) {

	bytes, err := json.Marshal(ids)
	if err != nil {
		return err
	}

	bundle.PackageIDs = string(bytes)
	return nil
}
