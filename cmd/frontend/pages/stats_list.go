package pages

import (
	"html/template"
	"math"
	"net/http"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/backend"
	"github.com/gamedb/gamedb/pkg/backend/generated"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func StatsListRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", statsListHandler)
	r.Get("/list.json", statsListJSONHandler)
	r.Mount("/{id}", StatRouter())
	return r
}

func statPathToConst(path string, r *http.Request) mongo.StatsType {

	switch path {
	case "categories":
		return mongo.StatsTypeCategories
	case "developers":
		return mongo.StatsTypeDevelopers
	case "genres":
		return mongo.StatsTypeGenres
	case "publishers":
		return mongo.StatsTypePublishers
	case "tags":
		return mongo.StatsTypeTags
	default:
		log.Warn("invalid stats type", zap.String("path", path), zap.String("path", r.URL.Path))
		return ""
	}

}

func statsListHandler(w http.ResponseWriter, r *http.Request) {

	typex := statPathToConst(chi.URLParam(r, "type"), r)

	t := statsTagsTemplate{}
	t.fill(w, r, "stats_list", typex.Title()+"s", template.HTML("Top Steam "+typex.Title()+"s"))
	t.addAssetMark()

	t.Type = typex

	returnTemplate(w, r, t)
}

type statsTagsTemplate struct {
	globalTemplate
	Type mongo.StatsType
}

func (t statsTagsTemplate) includes() []string {
	return []string{"includes/stats_header.gohtml"}
}

func statsListJSONHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	conn, ctx, err := backend.GetClient()
	if err != nil {
		log.ErrS(err)
		return
	}

	var columns = map[string]string{
		"0": "name",
		"1": "apps",
		"2": "mean_price",
		"3": "mean_score",
		"4": "mean_players",
	}

	code := session.GetProductCC(r)

	message := &generated.StatsRequest{
		Pagination: backend.MakePaginationRequest(query, columns, 100),
		Type:       query.GetSearchString("type"),
		Currency:   string(code),
		Search:     query.GetSearchString("search"),
	}

	resp, err := generated.NewStatsServiceClient(conn).List(ctx, message)
	if err != nil {
		log.ErrS(err)
		return
	}

	var response = datatable.NewDataTablesResponse(r, query, resp.GetPagination().GetTotal(), resp.GetPagination().GetTotalFiltered(), nil)
	for _, stat := range resp.GetStats() {

		statPath := helpers.GetStatPath(mongo.StatsType(query.GetSearchString("type")).MongoCol(), int(stat.GetId()), stat.GetName())
		statScore := helpers.GetAppReviewScore(float64(stat.GetMeanScore()))
		statPlayers := math.Round(float64(stat.GetMeanPlayers()))
		statPrice := i18n.FormatPrice(i18n.GetProdCC(code).CurrencyCode, int(math.Round(float64(stat.GetMeanPrice()))))

		response.AddRow([]interface{}{
			statPath,       // 0
			stat.GetName(), // 1
			stat.GetApps(), // 2
			statPrice,      // 3
			statPlayers,    // 4
			statScore,      // 5
		})
	}

	returnJSON(w, r, response)
}
