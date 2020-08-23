package pages

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gamedb/gamedb/cmd/frontend/pages/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/backend"
	"github.com/gamedb/gamedb/pkg/backend/generated"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
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
	var commits []*generated.CommitResponse
	var item = memcache.MemcacheCommitsPage(query.GetPage(commitsLimit))

	callback := func() (interface{}, error) {

		conn, ctx, err := backend.GetClient()
		if err != nil {
			return nil, err
		}

		message := &generated.CommitsRequest{
			Pagination: &generated.PaginationRequest2{
				Limit: commitsLimit,
				Page:  int64(query.GetPage(commitsLimit)),
			},
		}

		resp, err := generated.NewGitHubServiceClient(conn).Commits(ctx, message)
		if err != nil {
			return nil, err
		}

		return resp.GetCommits(), err
	}

	err := memcache.GetSetInterface(item.Key, item.Expiration, &commits, callback)
	if err != nil {
		zap.S().Error(err)
	}

	//
	total, err := strconv.Atoi(config.C.Commits)

	var response = datatable.NewDataTablesResponse(r, query, int64(total), int64(total), nil)
	var live bool

	for k, commit := range commits {

		t := time.Unix(commit.GetTime(), 0).Format(helpers.DateTime)

		if commit.GetHash() == config.C.CommitHash {
			live = true
		}

		response.AddRow([]interface{}{
			commit.GetMessage(),             // 0
			commit.GetTime(),                // 1
			t,                               // 2
			commit.GetLink(),                // 3
			commit.GetHash()[0:7],           // 4
			total - (query.GetOffset() + k), // 5
			live,                            // 6
		})
	}

	returnJSON(w, r, response)
}
