package db

import (
	"encoding/json"

	"github.com/gamedb/website/helpers"
)

type Bundle struct {
	ID         int
	Name       string
	Discount   int
	AppIDs     string
	PackageIDs string
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
