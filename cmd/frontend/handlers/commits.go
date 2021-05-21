package handlers

import (
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/backend"
	"github.com/gamedb/gamedb/pkg/backend/generated"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/ldflags"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/session"
	"github.com/go-chi/chi/v5"
)

func CommitsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", commitsHandler)
	r.Get("/commits.json", commitsAjaxHandler)
	return r
}

func commitsHandler(w http.ResponseWriter, r *http.Request) {

	t := commitsTemplate{}
	t.fill(w, r, "commits", "Commits", "All the open source commits for Global Steam")

	returnTemplate(w, r, t)
}

type commitsTemplate struct {
	globalTemplate
}

const commitsLimit = 100

func commitsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	var query = datatable.NewDataTableQuery(r, true)
	var commits []*generated.CommitResponse

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

	item := memcache.ItemCommitsPage(query.GetPage(commitsLimit))
	err := memcache.Client().GetSet(item.Key, item.Expiration, &commits, callback)
	if err != nil {
		log.ErrS(err)
	}

	//
	var commitsCount int
	if ldflags.CommitCount != "" {
		commitsCount, err = strconv.Atoi(ldflags.CommitCount)
		if err != nil {
			log.ErrS(err)
		}
	}

	var response = datatable.NewDataTablesResponse(r, query, int64(commitsCount), int64(commitsCount), nil)
	var live bool

	for _, commit := range commits {

		formattedTime := commit.GetTime().AsTime().Format(helpers.DateTime)

		if commit.GetHash() == ldflags.CommitHash {
			live = true
		}

		link := commit.GetLink()
		if session.IsAdmin(r) {
			link = "https://github.com/gamedb/gamedb/compare/" + commit.GetHash() + "...master?w=1"
		}

		response.AddRow([]interface{}{
			commit.GetMessage(),           // 0
			commit.GetTime().GetSeconds(), // 1
			formattedTime,                 // 2
			link,                          // 3
			commit.GetHash()[0:7],         // 4
			live,                          // 5
		})
	}

	returnJSON(w, r, response)
}
