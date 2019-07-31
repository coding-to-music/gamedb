package api

import (
	"github.com/gamedb/gamedb/pkg/sql"
)

type ApiApp struct {
	ID         int               `json:"id"`
	Name       string            `json:"name"`
	Tags       []int             `json:"tags"`
	Genres     []int             `json:"genres"`
	Developers []int             `json:"developers"`
	Publishers []int             `json:"publishers"`
	Prices     sql.ProductPrices `json:"prices"`
}

func (apiApp *ApiApp) Fill(sqlApp sql.App) (err error) {

	apiApp.ID = sqlApp.ID
	apiApp.Name = sqlApp.GetName()
	apiApp.Tags, err = sqlApp.GetTagIDs()
	if err != nil {
		return err
	}
	apiApp.Genres, err = sqlApp.GetGenreIDs()
	if err != nil {
		return err
	}
	apiApp.Developers, err = sqlApp.GetDeveloperIDs()
	if err != nil {
		return err
	}
	apiApp.Publishers, err = sqlApp.GetPublisherIDs()
	if err != nil {
		return err
	}
	apiApp.Prices, err = sqlApp.GetPrices()
	if err != nil {
		return err
	}

	return nil
}

