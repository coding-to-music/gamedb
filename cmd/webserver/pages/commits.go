package pages

import (
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/config"
	githubHelper "github.com/gamedb/gamedb/pkg/github"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/go-chi/chi"
	"github.com/google/go-github/v28/github"
)

func CommitsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", commitsHandler)
	r.Get("/commits.json", commitsAjaxHandler)
	return r
}

func commitsHandler(w http.ResponseWriter, r *http.Request) {

	t := commitsTemplate{}
	t.fill(w, r, "Commits", "All the open source commits for Game DB")

	returnTemplate(w, r, "commits", t)
}

type commitsTemplate struct {
	globalTemplate
}

const commitsLimit = 100

func commitsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	var query = datatable.NewDataTableQuery(r, true)
	var commits []*github.RepositoryCommit
	var item = memcache.MemcacheCommitsPage(query.GetPage(commitsLimit))

	err := memcache.GetSetInterface(item.Key, item.Expiration, &commits, func() (interface{}, error) {

		client, ctx := githubHelper.GetGithub()

		operation := func() (err error) {

			commits, _, err = client.Repositories.ListCommits(ctx, "gamedb", "website", &github.CommitsListOptions{
				ListOptions: github.ListOptions{
					Page:    query.GetPage(commitsLimit),
					PerPage: commitsLimit,
				},
			})
			return err
		}

		policy := backoff.NewExponentialBackOff()
		policy.InitialInterval = time.Second

		err := backoff.RetryNotify(operation, backoff.WithMaxRetries(policy, 5), func(err error, t time.Duration) { log.Info(err) })
		return commits, err
	})
	log.Err(err, r)

	//
	var total = config.Config.Commits.GetInt()
	var response = datatable.NewDataTablesResponse(r, query, int64(total), int64(total), nil)
	for k, commit := range commits {

		var date = commit.GetCommit().GetAuthor().GetDate().Format(helpers.DateTime)
		var unix = commit.GetCommit().GetAuthor().GetDate().Unix()
		var current = commit.GetSHA() == config.Config.CommitHash.Get()

		response.AddRow([]interface{}{
			commit.GetCommit().GetMessage(), // 0
			unix,                            // 1
			config.Config.Commits.GetInt(),  // 2
			commit.GetHTMLURL(),             // 3
			current,                         // 4
			commit.GetSHA()[0:7],            // 5
			date,                            // 6
			total - (query.GetOffset() + k), // 7
		})
	}

	returnJSON(w, r, response)
}
