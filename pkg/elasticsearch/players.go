package elasticsearch

import (
	"encoding/json"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steamid"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/olivere/elastic/v7"
)

type Player struct {
	ID                   int64                      `json:"id"`
	PersonaName          string                     `json:"name"`
	PersonaNameRecent    []string                   `json:"name_recent"`
	VanityURL            string                     `json:"url"`
	Avatar               string                     `json:"avatar"`
	Continent            string                     `json:"continent"`
	CountryCode          string                     `json:"country_code"`
	StateCode            string                     `json:"state_code"`
	LastBan              int64                      `json:"last_ban"`
	GameBans             int                        `json:"game_bans"`
	VACBans              int                        `json:"vac_bans"`
	Level                int                        `json:"level"`
	PlayTime             int                        `json:"play_time"`
	Badges               int                        `json:"badges"`
	BadgesFoil           int                        `json:"badges_foil"`
	Games                int                        `json:"games"`
	Achievements         int                        `json:"achievements"`
	Achievements100      int                        `json:"achievements_100"`
	AwardsGivenCount     int                        `json:"awards_given_count"`
	AwardsGivenPoints    int                        `json:"awards_given_points"`
	AwardsReceivedCount  int                        `json:"awards_received_count"`
	AwardsReceivedPoints int                        `json:"awards_received_points"`
	Ranks                map[helpers.RankMetric]int `json:"ranks"`
	PersonaNameMarked    string                     `json:"-"`
	Score                float64                    `json:"-"`
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

func (player Player) GetPathAbsolute() string {
	return helpers.GetPlayerPathAbsolute(player.ID, player.PersonaName)
}

func (player Player) GetAvatar() string {
	return helpers.GetPlayerAvatar(player.Avatar)
}

func (player Player) GetAvatarAbsolute() string {
	return helpers.GetPlayerAvatarAbsolute(player.Avatar)
}

func (player Player) GetAvatar2() string {
	return helpers.GetPlayerAvatar2(player.Level)
}

func (player Player) GetFlag() string {
	return helpers.GetPlayerFlagPath(player.CountryCode)
}

func (player Player) GetCountry() string {
	return i18n.CountryCodeToName(player.CountryCode)
}

func (player Player) GetCommunityLink() string {
	return helpers.GetPlayerCommunityLink(player.ID, player.VanityURL)
}

func (player Player) GetGamesCount() int {
	return player.Games
}

func (player Player) GetAchievements() int {
	return player.Achievements
}

func (player Player) GetPlaytime() int {
	return player.PlayTime
}

func (player Player) GetLevel() int {
	return player.Level
}

func (player Player) GetBadges() int {
	return player.Badges
}

func (player Player) GetBadgesFoil() int {
	return player.BadgesFoil
}

func (player Player) GetRanks() map[helpers.RankMetric]int {
	return player.Ranks
}

func (player Player) GetVACBans() int {
	return player.VACBans
}

func (player Player) GetGameBans() int {
	return player.GameBans
}

func (player Player) GetLastBan() time.Time {
	return time.Unix(player.LastBan, 0)
}

func IndexPlayer(p Player) error {
	return indexDocument(IndexPlayers, strconv.FormatInt(p.ID, 10), p)
}

func SearchPlayers(limit int, offset int, search string, sorters []elastic.Sorter, filters []elastic.Query) (players []Player, total int64, err error) {

	client, ctx, err := client()
	if err != nil {
		return players, 0, err
	}

	searchService := client.Search().
		Index(IndexPlayers).
		From(offset).
		Size(limit).
		TrackTotalHits(true)

	var query = elastic.NewBoolQuery().Filter(filters...)

	if search != "" {

		var musts []elastic.Query

		if strings.HasPrefix(search, "http") {

			search = path.Base(search)

			playerID, err := steamid.ParsePlayerID(search)
			if err != nil {
				musts = []elastic.Query{elastic.NewTermQuery("url", search).Boost(2)}
			} else {
				musts = []elastic.Query{elastic.NewTermQuery("id", playerID).Boost(6)}
			}

		} else {

			musts = []elastic.Query{
				elastic.NewTermQuery("id", search).Boost(6),
				elastic.NewMatchQuery("name", search).Boost(4),
				elastic.NewTermQuery("url", search).Boost(2),
			}

			query.Should(
				elastic.NewFunctionScoreQuery().AddScoreFunc(elastic.NewFieldValueFactorFunction().
					Modifier("sqrt").Field("level").Factor(0.1)),
				elastic.NewFunctionScoreQuery().AddScoreFunc(elastic.NewFieldValueFactorFunction().
					Modifier("sqrt").Field("games").Factor(0.1)),
			)
		}

		query.Must(
			elastic.NewBoolQuery().MinimumNumberShouldMatch(1).Should(musts...),
		)

		searchService.Highlight(elastic.NewHighlight().Field("name").PreTags("<mark>").PostTags("</mark>"))

	} else {
		if len(sorters) > 0 {
			searchService.SortBy(sorters...)
		}
	}

	searchService.Query(query)

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return players, 0, err
	}

	for _, hit := range searchResult.Hits.Hits {

		var player Player
		err := json.Unmarshal(hit.Source, &player)
		if err != nil {
			log.ErrS(err)
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

	return players, searchResult.TotalHits(), err
}

func AggregatePlayerCountries() (aggregations map[string]int64, err error) {

	item := memcache.ItemPlayerLocationAggs
	err = memcache.Client().GetSet(item.Key, item.Expiration, &aggregations, func() (interface{}, error) {

		client, ctx, err := client()
		if err != nil {
			return aggregations, err
		}

		searchService := client.Search().
			Index(IndexPlayers).
			Aggregation("country", elastic.NewTermsAggregation().Field("country_code").Size(1000).
				SubAggregation("state", elastic.NewTermsAggregation().Field("state_code").Size(1000)),
			).
			Aggregation("continent", elastic.NewTermsAggregation().Field("continent").Size(10))

		searchResult, err := searchService.Do(ctx)
		if err != nil {
			return aggregations, err
		}

		aggregations = map[string]int64{}

		if a, ok := searchResult.Aggregations.Terms("country"); ok {
			for _, country := range a.Buckets {
				aggregations[country.Key.(string)] = country.DocCount
				if a, ok := country.Terms("state"); ok {
					for _, state := range a.Buckets {
						aggregations[country.Key.(string)+"-"+state.Key.(string)] = state.DocCount
					}
				}
			}
		}

		if a, ok := searchResult.Aggregations.Terms("continent"); ok {
			for _, country := range a.Buckets {
				aggregations["c-"+country.Key.(string)] = country.DocCount
			}
		}

		return aggregations, err
	})

	return aggregations, err
}

//noinspection GoUnusedExportedFunction
func DeleteAndRebuildPlayersIndex() {

	var rankProperties = map[string]interface{}{}
	for _, v := range helpers.PlayerRankFields {
		rankProperties[string(v)] = fieldTypeInt32
	}

	var mapping = map[string]interface{}{
		"settings": settings,
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":                     fieldTypeKeyword,
				"name":                   fieldTypeText,
				"name_recent":            fieldTypeText,
				"url":                    fieldTypeText,
				"avatar":                 fieldTypeDisabled,
				"continent":              fieldTypeKeyword,
				"country_code":           fieldTypeKeyword,
				"state_code":             fieldTypeKeyword,
				"last_ban":               fieldTypeInt64,
				"game_bans":              fieldTypeInt32,
				"vac_bans":               fieldTypeInt32,
				"level":                  fieldTypeInt32,
				"play_time":              fieldTypeInt32,
				"badges":                 fieldTypeInt32,
				"badges_foil":            fieldTypeInt32,
				"games":                  fieldTypeInt32,
				"achievements":           fieldTypeInt32,
				"achievements_100":       fieldTypeInt32,
				"awards_given_count":     fieldTypeInt32,
				"awards_given_points":    fieldTypeInt32,
				"awards_received_count":  fieldTypeInt32,
				"awards_received_points": fieldTypeInt32,
				"ranks":                  fieldTypeDisabled,
			},
		},
	}

	rebuildIndex(IndexPlayers, mapping)
}
