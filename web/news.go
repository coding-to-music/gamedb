package web

import (
	"net/http"

	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/logging"
)

func NewsHandler(w http.ResponseWriter, r *http.Request) {

	articles, err := db.GetArticles(0, 100)
	if err != nil {
		logging.Error(err)
		returnErrorTemplate(w, r, 500, "Error getting articles")
		return
	}

	// Make template articles
	var appIDs []int
	var templateArticles []newsArticleTemplate
	for _, v := range articles {

		if v.AppID != 0 {

			templateArticles = append(templateArticles, newsArticleTemplate{
				Article: v,
			})

			appIDs = append(appIDs, v.AppID)
		}
	}

	// Get apps
	apps, err := db.GetAppsByID(appIDs, []string{"id", "name", "icon"})
	if err != nil {
		logging.Error(err)
		returnErrorTemplate(w, r, 500, "Error getting apps")
		return
	}

	// Make map of apps
	var appsMap = map[int]db.App{}
	for _, v := range apps {
		appsMap[v.ID] = v
	}

	// Add apps to template
	for k, v := range templateArticles {

		if val, ok := appsMap[v.Article.AppID]; ok {
			templateArticles[k].App = val
		}
	}

	// Template
	t := newsTemplate{}
	t.Fill(w, r, "News")
	t.Articles = templateArticles

	returnTemplate(w, r, "news", t)
	return
}

type newsTemplate struct {
	GlobalTemplate
	Articles []newsArticleTemplate
}

type newsArticleTemplate struct {
	Article db.News
	App     db.App
}
