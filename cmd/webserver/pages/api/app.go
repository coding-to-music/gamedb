package api

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
)

type App struct {
	ID         int               `json:"id"`
	Name       string            `json:"name"`
	Tags       []int             `json:"tags"`
	Genres     []int             `json:"genres"`
	Developers []int             `json:"developers"`
	Publishers []int             `json:"publishers"`
	Prices     sql.ProductPrices `json:"prices"`
}

func (apiApp *App) Fill(sqlApp sql.App) (err error) {

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

func ApiAppsHandler(call APIRequest) (ret interface{}, err error) {

	//noinspection GoPreferNilSlice
	apps := []App{}

	//
	db, err := sql.GetMySQLClient()
	if err != nil {
		return apps, err
	}

	db = db.Select([]string{"id", "name", "tags", "genres", "developers", "categories", "prices"})
	db, err = call.SetSQLLimitOffset(db)
	if err != nil {
		return apps, err
	}

	var sqlApps []sql.App
	db = db.Find(&sqlApps)
	if db.Error != nil {
		return apps, err
	}

	for _, v := range sqlApps {
		apiApp := App{}
		err = apiApp.Fill(v)
		log.Err(err)

		apps = append(apps, apiApp)
	}

	return apps, nil
}
