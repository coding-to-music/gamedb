package pages

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
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

	if strings.Contains(r.URL.RawQuery, "refresh") {
		err := memcache.Delete(memcache.MemcacheCommitsPage(1).Key)
		log.Err(err)
		http.Redirect(w, r, "/commits", http.StatusFound)
		return
	}

	returnTemplate(w, r, "commits", t)
}

type commitsTemplate struct {
	GlobalTemplate
}

const commitsLimit = 100

func commitsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, true)

	client, ctx := helpers.GetGithub()

	var wg sync.WaitGroup

	var commits []*github.RepositoryCommit
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		var item = memcache.MemcacheCommitsPage(query.GetPage(commitsLimit))

		err = memcache.GetSetInterface(item.Key, item.Expiration, &commits, func() (interface{}, error) {

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
	}()

	var total int
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		var item = memcache.MemcacheCommitsTotal

		err = memcache.GetSetInterface(item.Key, item.Expiration, &total, func() (interface{}, error) {

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
			policy.InitialInterval = time.Second

			err := backoff.RetryNotify(operation, backoff.WithMaxRetries(policy, 5), func(err error, t time.Duration) { log.Info(err) })
			return total, err
		})
		log.Err(err, r)
	}()

	wg.Wait()

	//
	var deployed bool
	var response = datatable.NewDataTablesResponse(r, query, int64(total), int64(total))
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

	returnJSON(w, r, response)
}
