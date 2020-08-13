package pages

import (
	"io"
	"net/http"
	"time"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/backend"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/go-chi/chi"
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
	var commits []backend.CommitResponse
	var item = memcache.MemcacheCommitsPage(query.GetPage(commitsLimit))

	err := memcache.GetSetInterface(item.Key, item.Expiration, &commits, func() (interface{}, error) {

		conn, ctx, err := backend.GetClient()
		if err != nil {
			return nil, err
		}

		message := &backend.CommitsRequest{
			Limit: commitsLimit,
			Page:  int32(query.GetPage(commitsLimit)),
		}

		resp, err := backend.NewGitHubServiceClient(conn).Commits(ctx, message)
		if err != nil {
			return nil, err
		}

		for {

			commit, err := resp.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Err(err, r)
				continue
			}

			commits = append(commits, *commit)
		}

		return commits, err
	})
	log.Err(err, r)

	//
	var total = config.Config.Commits.GetInt()
	var response = datatable.NewDataTablesResponse(r, query, int64(total), int64(total), nil)

	for k, commit := range commits {

		t := time.Unix(commit.GetTime(), 0).Format(helpers.DateTime)

		response.AddRow([]interface{}{
			commit.GetMessage(),             // 0
			commit.GetTime(),                // 1
			t,                               // 2
			commit.GetLink(),                // 3
			commit.GetHash()[0:7],           // 4
			total - (query.GetOffset() + k), // 5
			config.Config.Commits.GetInt(),  // 6
		})
	}

	returnJSON(w, r, response)
}
