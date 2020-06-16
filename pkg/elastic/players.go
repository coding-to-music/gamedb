package elastic

import (
	"encoding/json"
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/olivere/elastic/v7"
)

type Player struct {
	ID                int64    `json:"id"`
	PersonaName       string   `json:"name"`
	PersonaNameMarked string   `json:"name_marked"`
	PersonaNameRecent []string `json:"name_recent"`
	VanityURL         string   `json:"url"`
	Avatar            string   `json:"avatar"`
	CountryCode       string   `json:"country_code"`
	StateCode         string   `json:"state_code"`
	LastBan           int64    `json:"last_ban"`
	GameBans          int      `json:"game_bans"`
	VACBans           int      `json:"vac_bans"`
	Level             int      `json:"level"`
	PlayTime          int      `json:"play_time"`
	Badges            int      `json:"badges"`
	Games             int      `json:"games"`
	Friends           int      `json:"friends"`
	Comments          int      `json:"comments"`
	Score             float64  `json:"-"`
}

func (player Player) GetName() string {
	return helpers.GetPlayerName(player.ID, player.PersonaName)
}

func (player Player) GetNameMarked() string {
	return helpers.GetPlayerName(player.ID, player.PersonaNameMarked)
}

func (player Player) GetPath() string {
	return helpers.GetPlayerPath(player.ID, player.PersonaName)
}

func (player Player) GetAvatar() string {
	return helpers.GetPlayerAvatar(player.Avatar)
}

func (player Player) GetCommunityLink() string {
	return helpers.GetPlayerCommunityLink(player.ID, player.VanityURL)
}

func IndexPlayer(player Player) error {
	return indexDocument(IndexPlayers, strconv.FormatInt(player.ID, 10), player)
}

func SearchPlayers(limit int, offset int, search string, sorters []elastic.Sorter) (players []Player, aggregations map[string]map[string]int64, total int64, err error) {

	client, ctx, err := GetElastic()
	if err != nil {
		return players, aggregations, 0, err
	}

	searchService := client.Search().
		Index(IndexPlayers).
		From(offset).
		Size(limit).
		TrackTotalHits(true).
		Highlight(elastic.NewHighlight().Field("name").PreTags("<mark>").PostTags("</mark>")).
		Aggregation("country", elastic.NewTermsAggregation().Field("type").Size(10).OrderByCountDesc().
			SubAggregation("state", elastic.NewTermsAggregation().Field("state_code").Size(10).OrderByCountDesc()),
		)

	if search != "" {

		var search2 = helpers.RegexNonAlphaNumeric.ReplaceAllString(search, "")

		searchService.Query(elastic.NewBoolQuery().
			Must(
				elastic.NewBoolQuery().MinimumNumberShouldMatch(1).Should(
					elastic.NewTermQuery("id", search).Boost(5),
					elastic.NewTermQuery("url", search2).Boost(1),
					elastic.NewMatchQuery("name", search).Boost(2),
					elastic.NewMatchQuery("name_recent", search).Boost(1),
				),
			).
			Should(
				elastic.NewFunctionScoreQuery().
					AddScoreFunc(elastic.NewFieldValueFactorFunction().Modifier("sqrt").Field("level").Factor(0.01)).
					AddScoreFunc(elastic.NewFieldValueFactorFunction().Modifier("sqrt").Field("games").Factor(0.001)),
			),
		)
	}

	if len(sorters) > 0 {
		searchService.SortBy(sorters...)
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return players, aggregations, 0, err
	}

	aggregations = make(map[string]map[string]int64, len(searchResult.Aggregations))
	for k := range searchResult.Aggregations {
		if a, ok := searchResult.Aggregations.Terms(k); ok {
			aggregations[k] = make(map[string]int64, len(a.Buckets))
			for _, v := range a.Buckets {
				aggregations[k][*v.KeyAsString] = v.DocCount
			}
		}
	}

	for _, hit := range searchResult.Hits.Hits {

		var player Player
		err := json.Unmarshal(hit.Source, &player)
		if err != nil {
			log.Err(err)
		}

		if hit.Score != nil {
			player.Score = *hit.Score
		}

		player.PersonaNameMarked = player.PersonaName
		if val, ok := hit.Highlight["name"]; ok {
			if len(val) > 0 {
				player.PersonaNameMarked = val[0]
			}
		}

		players = append(players, player)
	}

	return players, aggregations, searchResult.TotalHits(), err
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildPlayersIndex() {

	var mapping = map[string]interface{}{
		"settings": settings,
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":           fieldTypeKeyword,
				"name":         fieldTypeText,
				"name_recent":  fieldTypeText,
				"url":          fieldTypeText,
				"avatar":       fieldTypeDisabled,
				"country_code": fieldTypeKeyword,
				"state_code":   fieldTypeKeyword,
				"last_ban":     fieldTypeLong,
				"game_bans":    fieldTypeInteger,
				"vac_bans":     fieldTypeInteger,
				"level":        fieldTypeInteger,
				"play_time":    fieldTypeInteger,
				"badges":       fieldTypeInteger,
				"games":        fieldTypeInteger,
				"friends":      fieldTypeInteger,
				"comments":     fieldTypeInteger,
			},
		},
	}

	rebuildIndex(IndexPlayers, mapping)
}
