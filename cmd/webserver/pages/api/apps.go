package api

import (
	"errors"
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
)

type App struct {
	ID              int                   `json:"id"`
	Name            string                `json:"name"`
	Tags            []int                 `json:"tags"`
	Genres          []int                 `json:"genres"`
	Categories      []int                 `json:"categories"`
	Developers      []int                 `json:"developers"`
	Publishers      []int                 `json:"publishers"`
	Prices          helpers.ProductPrices `json:"prices"`
	PlayersMax      int                   `json:"players_max"`
	PlayersWeekMax  int                   `json:"players_week_max"`
	PlayersWeekAvg  float64               `json:"players_week_avg"`
	ReleaseDate     int64                 `json:"release_date"`
	ReviewsPositive int                   `json:"reviews_positive"`
	ReviewsNegative int                   `json:"reviews_negative"`
	ReviewsScore    float64               `json:"reviews_score"`
}

func (app *App) Fill(sqlApp sql.App) (err error) {

	app.ID = sqlApp.ID
	app.Name = sqlApp.GetName()
	app.Tags = sqlApp.GetTagIDs()
	app.Genres = sqlApp.GetGenreIDs()
	app.Developers = sqlApp.GetDeveloperIDs()
	app.Categories = sqlApp.GetCategoryIDs()
	app.Publishers = sqlApp.GetPublisherIDs()
	app.Prices = sqlApp.GetPrices()
	app.PlayersMax = sqlApp.PlayerPeakAllTime
	app.PlayersWeekMax = sqlApp.PlayerPeakWeek
	app.PlayersWeekAvg = sqlApp.PlayerAverageWeek
	app.ReleaseDate = sqlApp.ReleaseDateUnix
	app.ReviewsScore = sqlApp.ReviewsScore

	reviews := sqlApp.GetReviews()
	app.ReviewsPositive = reviews.Positive
	app.ReviewsNegative = reviews.Negative

	return nil
}

func AppsHandler(call APIRequest) (ret interface{}, err error) {

	//noinspection GoPreferNilSlice
	apps := []App{}

	db, err := sql.GetMySQLClient()
	if err != nil {
		return apps, err
	}

	// Select columns
	db, err = call.setSQLSelect(db, []string{
		"id",
		"name",
		"tags",
		"genres",
		"developers",
		"categories",
		"prices",
		"player_peak_alltime",
		"player_peak_week",
		"player_avg_week",
		"release_date_unix",
		"reviews",
		"reviews_score",
	})
	if err != nil {
		return apps, err
	}

	// Limit & Offset (page)
	db, err = call.setSQLLimitOffset(db)
	if err != nil {
		return apps, err
	}

	// Order field & order
	db, err = call.setSQLOrder(db, func(in string) (out string) {
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
	})
	if err != nil {
		return apps, err
	}

	// ID
	i, err := call.getQueryInt("id")
	if err != errParamNotSet {
		if err != nil {
			return apps, err
		}
		if i < 1 {
			return apps, errors.New("invalid id")
		}
		db = db.Where("id = ?", i)
	}

	// Tag
	i, err = call.getQueryInt("tag")
	if err != errParamNotSet {
		if err != nil {
			return apps, err
		}
		if i < 1 {
			return apps, errors.New("invalid tag")
		}
		db = db.Where("JSON_CONTAINS(tags, ?) = 1", "["+strconv.FormatInt(i, 10)+"]")
	}

	// Category
	i, err = call.getQueryInt("category")
	if err != errParamNotSet {
		if err != nil {
			return apps, err
		}
		if i < 1 {
			return apps, errors.New("invalid category")
		}
		db = db.Where("JSON_CONTAINS(categories, ?) = 1", "["+strconv.FormatInt(i, 10)+"]")
	}

	// Genre
	i, err = call.getQueryInt("genre")
	if err != errParamNotSet {
		if err != nil {
			return apps, err
		}
		if i < 1 {
			return apps, errors.New("invalid genre")
		}
		db = db.Where("JSON_CONTAINS(genres, ?) = 1", "["+strconv.FormatInt(i, 10)+"]")
	}

	// Min players
	i, err = call.getQueryInt("min_players")
	if err != errParamNotSet {
		if err != nil {
			return apps, err
		}
		if i < 0 {
			return apps, errors.New("invalid min players")
		}
		db = db.Where("player_peak_week >= ?", i)
	}

	// Max players
	i, err = call.getQueryInt("max_players")
	if err != errParamNotSet {
		if err != nil {
			return apps, err
		}
		if i < 1 {
			return apps, errors.New("invalid max players")
		}
		db = db.Where("player_peak_week <= ?", i)
	}

	// Min avg players
	i, err = call.getQueryInt("min_avg_players")
	if err != errParamNotSet {
		if err != nil {
			return apps, err
		}
		if i < 0 {
			return apps, errors.New("invalid min avg players")
		}
		db = db.Where("player_avg_week >= ?", i)
	}

	// Max avg players
	i, err = call.getQueryInt("max_avg_players")
	if err != errParamNotSet {
		if err != nil {
			return apps, err
		}
		if i < 1 {
			return apps, errors.New("invalid max avg players")
		}
		db = db.Where("player_avg_week <= ?", i)
	}

	// Min release date
	i, err = call.getQueryInt("min_release_date")
	if err != errParamNotSet {
		if err != nil {
			return apps, err
		}
		if i < 0 {
			return apps, errors.New("invalid release date")
		}
		db = db.Where("release_date_unix >= ?", i)
	}

	// Max release date
	i, err = call.getQueryInt("max_release_date")
	if err != errParamNotSet {
		if err != nil {
			return apps, err
		}
		if i < 1 {
			return apps, errors.New("invalid release date")
		}
		db = db.Where("release_date_unix <= ?", i)
	}

	// Min review score
	i, err = call.getQueryInt("min_score")
	if err != errParamNotSet {
		if err != nil {
			return apps, err
		}
		if i < 0 {
			return apps, errors.New("invalid review score")
		}
		db = db.Where("reviews_score >= ?", i)
	}

	// Max review score
	i, err = call.getQueryInt("max_score")
	if err != errParamNotSet {
		if err != nil {
			return apps, err
		}
		if i < 1 {
			return apps, errors.New("invalid review score")
		}
		db = db.Where("reviews_score <= ?", i)
	}

	// Min trending value
	i, err = call.getQueryInt("min_trending")
	if err != errParamNotSet {
		if err != nil {
			return apps, err
		}
		if i < 0 {
			return apps, errors.New("invalid trending value")
		}
		db = db.Where("player_trend >= ?", i)
	}

	// Max trending value
	i, err = call.getQueryInt("max_trending")
	if err != errParamNotSet {
		if err != nil {
			return apps, err
		}
		if i < 1 {
			return apps, errors.New("invalid trending value")
		}
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
