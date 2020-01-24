package pages

import (
	"errors"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/cmd/webserver/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	githubHelper "github.com/gamedb/gamedb/pkg/helpers/github"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
	"github.com/google/go-github/v28/github"
)

const (
	commitsLimit = 100
)

func CommitsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", commitsHandler)
	r.Get("/commits.json", commitsAjaxHandler)
	return r
}

func commitsHandler(w http.ResponseWriter, r *http.Request) {

	t := commitsTemplate{}
	t.fill(w, r, "Commits", "")

	t.Total = getTotalCommits()

	returnTemplate(w, r, "commits", t)
}

type commitsTemplate struct {
	GlobalTemplate
	Total int
}

func commitsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, true)

	client, ctx := githubHelper.GetGithub()

	commits, _, err := client.Repositories.ListCommits(ctx, "gamedb", "website", &github.CommitsListOptions{
		ListOptions: github.ListOptions{
			Page:    query.GetPage(commitsLimit),
			PerPage: commitsLimit,
		},
	})

	if err != nil {
		log.Err(err, r)
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the commits."})
		return
	}

	// Get total
	total := getTotalCommits()

	//
	response := datatable.DataTablesResponse{}
	response.RecordsTotal = int64(total)
	response.RecordsFiltered = int64(total)
	response.Draw = query.Draw
	response.Limit(r)

	var deployed bool
	for _, commit := range commits {

		if commit.GetSHA() == config.Config.CommitHash.Get() {
			deployed = true
		}

		response.AddRow([]interface{}{
			commit.GetCommit().GetMessage(),                 // 0
			commit.GetCommit().GetAuthor().GetDate().Unix(), // 1
			deployed,            // 2
			commit.GetHTMLURL(), // 3
			commit.GetSHA() == config.Config.CommitHash.Get(),                 // 4
			commit.GetSHA()[0:7],                                              // 5
			commit.GetCommit().GetAuthor().GetDate().Format(helpers.DateTime), // 6
		})
	}

	returnJSON(w, r, response.Output())
}

func getTotalCommits() (total int) {

	client, ctx := githubHelper.GetGithub()

	var item = memcache.MemcacheTotalCommits

	err := memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &total, func() (interface{}, error) {

		operation := func() (err error) {

			contributors, _, err := client.Repositories.ListContributorsStats(ctx, "gamedb", "gamedb")
			if err != nil { // github.AcceptedError
				return err
			}
			for _, v := range contributors {
				total += v.GetTotal()
			}
			if total == 0 {
				return errors.New("no commits found")
			}
			return nil
		}

		policy := backoff.NewExponentialBackOff()
		policy.InitialInterval = time.Second / 2

		err := backoff.RetryNotify(operation, backoff.WithMaxRetries(policy, 4), func(err error, t time.Duration) { log.Info(err) })
		return total, err
	})
	log.Err(err)

	return total
}
