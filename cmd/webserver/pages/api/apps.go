package api

import (
	"errors"
	"strconv"

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

	db, err = call.setSQLLimitOffset(db)
	if err != nil {
		return apps, err
	}

	db, err = call.setSQLOrder(db, mutateAppsOrder)
	if err != nil {
		return apps, err
	}

	// ID
	i, err := call.getQueryInt("id", 0)
	if err != nil {
		return apps, errors.New("invalid id")
	}
	if i > 0 {
		db = db.Where("id = ?", i)
	}

	// Tag
	i, err = call.getQueryInt("tag", 0)
	if err != nil {
		return apps, errors.New("invalid tag")
	}
	if i > 0 {
		db = db.Where("JSON_CONTAINS(tags, ?) = 1", "["+strconv.FormatInt(i, 10)+"]")
	}

	// Category
	i, err = call.getQueryInt("category", 0)
	if err != nil {
		return apps, errors.New("invalid category")
	}
	if i > 0 {
		db = db.Where("JSON_CONTAINS(categories, ?) = 1", "["+strconv.FormatInt(i, 10)+"]")
	}

	// Genre
	i, err = call.getQueryInt("genre", 0)
	if err != nil {
		return apps, errors.New("invalid genre")
	}
	if i > 0 {
		db = db.Where("JSON_CONTAINS(genres, ?) = 1", "["+strconv.FormatInt(i, 10)+"]")
	}

	// Min players
	i, err = call.getQueryInt("min_players", 0)
	if err != nil {
		return apps, errors.New("invalid min players")
	}
	if i >= 0 {
		db = db.Where("player_peak_week >= ?", i)
	}

	// Max players
	i, err = call.getQueryInt("max_players", 0)
	if err != nil {
		return apps, errors.New("invalid max players")
	}
	if i >= 0 {
		db = db.Where("player_peak_week <= ?", i)
	}

	// Min release date
	i, err = call.getQueryInt("min_release_date", 0)
	if err != nil {
		return apps, errors.New("invalid release date")
	}
	if i > 0 {
		db = db.Where("release_date_unix >= ?", i)
	}

	// Max release date
	i, err = call.getQueryInt("max_release_date", 0)
	if err != nil {
		return apps, errors.New("invalid release date")
	}
	if i > 0 {
		db = db.Where("release_date_unix <= ?", i)
	}

	// Min review score
	i, err = call.getQueryInt("min_score", 0)
	if err != nil {
		return apps, errors.New("invalid review score")
	}
	if i > 0 {
		db = db.Where("reviews_score >= ?", i)
	}

	// Max review score
	i, err = call.getQueryInt("max_score", 0)
	if err != nil {
		return apps, errors.New("invalid review score")
	}
	if i > 0 {
		db = db.Where("reviews_score <= ?", i)
	}

	// Min trending value
	i, err = call.getQueryInt("min_trending", 0)
	if err != nil {
		return apps, errors.New("invalid trending value")
	}
	if i > 0 {
		db = db.Where("player_trend >= ?", i)
	}

	// Max trending value
	i, err = call.getQueryInt("max_trending", 0)
	if err != nil {
		return apps, errors.New("invalid trending value")
	}
	if i > 0 {
		db = db.Where("player_trend <= ?", i)
	}

	//
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

func mutateAppsOrder(in string) (out string) {

	switch in {
	case "id", "name":
		return in
	case "players":
		return "player_peak_week"
	case "release_date":
		return "release_date_unix"
	default:
		return ""
	}
}
