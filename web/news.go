package web

import (
	"net/http"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func NewsHandler(w http.ResponseWriter, r *http.Request) {

	articles, err := datastore.GetArticles(0, 100)
	if err != nil {
		logger.Error(err)
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
	apps, err := mysql.GetApps(appIDs, []string{"id", "name", "icon"})
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, "Error getting apps")
		return
	}

	// Make map of apps
	var appsMap = map[int]mysql.App{}
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
	template := newsTemplate{}
	template.Fill(r, "News")
	template.Articles = templateArticles

	returnTemplate(w, r, "news", template)
	return
}

type newsTemplate struct {
	GlobalTemplate
	Articles []newsArticleTemplate
}

type newsArticleTemplate struct {
	Article datastore.Article
	App     mysql.App
}
