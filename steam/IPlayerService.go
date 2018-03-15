package steam

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/kr/pretty"
)

// Gets information about a player's recently played games
func GetRecentlyPlayedGames(playerID int) (games []RecentlyPlayedGame, err error) {

	options := url.Values{}
	options.Set("steamid", strconv.Itoa(playerID))
	options.Set("count", "0")

	bytes, err := get("IPlayerService/GetRecentlyPlayedGames/v1", options)
	if err != nil {
		return games, err
	}

	// Unmarshal JSON
	var resp RecentlyPlayedGamesResponse
	err = json.Unmarshal(bytes, &resp)
	if err != nil {
		pretty.Print(err.Error())
		pretty.Print(string(bytes))
		return games, err
	}

	return resp.Response.Games, nil
}

type RecentlyPlayedGamesResponse struct {
	Response struct {
		TotalCount int                  `json:"total_count"`
		Games      []RecentlyPlayedGame `json:"games"`
	} `json:"response"`
}

type RecentlyPlayedGame struct {
	AppID           int    `json:"appid"`
	Name            string `json:"name"`
	Playtime2Weeks  int    `json:"playtime_2weeks"`
	PlaytimeForever int    `json:"playtime_forever"`
	ImgIconURL      string `json:"img_icon_url"`
	ImgLogoURL      string `json:"img_logo_url"`
}

// Return a list of games owned by the player
func GetOwnedGames(id int) (games []OwnedGame, err error) {

	options := url.Values{}
	options.Set("steamid", strconv.Itoa(id))
	options.Set("include_appinfo", "1")
	options.Set("include_played_free_games", "1")

	bytes, err := get("IPlayerService/GetOwnedGames/v1", options)
	if err != nil {
		return games, err
	}

	// Unmarshal JSON
	var resp OwnedGamesResponse
	err = json.Unmarshal(bytes, &resp)
	if err != nil {
		return games, err
	}

	return resp.Response.Games, nil
}

type OwnedGamesResponse struct {
	Response struct {
		GameCount int         `json:"game_count"`
		Games     []OwnedGame `json:"games"`
	} `json:"response"`
}

type OwnedGame struct {
	AppID                    int    `json:"appid"`
	Name                     string `json:"name"`
	PlaytimeForever          int    `json:"playtime_forever"`
	ImgIconURL               string `json:"img_icon_url"`
	ImgLogoURL               string `json:"img_logo_url"`
	HasCommunityVisibleStats bool   `json:"has_community_visible_stats,omitempty"`
}

// Returns the Steam Level of a user
func GetSteamLevel(id int) (level int, err error) {

	options := url.Values{}
	options.Set("steamid", strconv.Itoa(id))

	bytes, err := get("IPlayerService/GetSteamLevel/v1", options)
	if err != nil {
		return level, err
	}

	// Unmarshal JSON
	var resp LevelResponse
	err = json.Unmarshal(bytes, &resp)
	if err != nil {
		return level, err
	}

	return resp.Response.PlayerLevel, nil
}

type LevelResponse struct {
	Response struct {
		PlayerLevel int `json:"player_level"`
	} `json:"response"`
}

// Gets badges that are owned by a specific user
func GetBadges(id int) (badges BadgesResponse, err error) {

	options := url.Values{}
	options.Set("steamid", strconv.Itoa(id))

	bytes, err := get("IPlayerService/GetBadges/v1", options)
	if err != nil {
		return badges, err
	}

	// Unmarshal JSON
	var resp BadgesResponseOuter
	err = json.Unmarshal(bytes, &resp)
	if err != nil {
		return badges, err
	}

	return resp.Response, nil
}

type BadgesResponseOuter struct {
	Response BadgesResponse `json:"response"`
}

type BadgesResponse struct {
	Badges []struct {
		Badgeid        int `json:"badgeid"`
		Level          int `json:"level"`
		CompletionTime int `json:"completion_time"`
		Xp             int `json:"xp"`
		Scarcity       int `json:"scarcity"`
	} `json:"badges"`
	PlayerXP                   int `json:"player_xp"`
	PlayerLevel                int `json:"player_level"`
	PlayerXPNeededToLevelUp    int `json:"player_xp_needed_to_level_up"`
	PlayerXPNeededCurrentLevel int `json:"player_xp_needed_current_level"`
}
