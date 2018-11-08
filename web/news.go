package web

import (
	"net/http"

	"github.com/gamedb/website/db"
)

func NewsHandler(w http.ResponseWriter, r *http.Request) {

	articles, err := db.GetArticles()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Error getting articles.", Error: err})
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
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Error getting apps.", Error: err})
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
