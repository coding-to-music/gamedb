package utils

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/oauth"
)

type migrateOAuths struct{}

func (migrateOAuths) name() string {
	return "migrateOAuths"
}

func (migrateOAuths) run() {

	db, err := mysql.GetMySQLClient()
	if err != nil {
		log.ErrS(err)
		return
	}

	var users []mysql.User

	db = db.Model(&mysql.User{})
	db = db.Select([]string{"id", "discord_id", "github_id", "google_id", "patreon_id", "steam_id"})
	db = db.Limit(1000)
	db = db.Offset(0)
	db = db.Order("id asc")
	db = db.Find(&users)

	if db.Error != nil {
		log.ErrS(db.Error)
		return
	}

	for _, v := range users {

		log.InfoS(v.ID)

		if v.DiscordID.Valid && v.DiscordID.String != "" {
			createUser(v.ID, oauth.ProviderDiscord, v.DiscordID.String)
		}

		if v.GitHubID.Valid && v.GitHubID.String != "" {
			createUser(v.ID, oauth.ProviderGithub, v.GitHubID.String)
		}

		if v.GoogleID.Valid && v.GoogleID.String != "" {
			createUser(v.ID, oauth.ProviderGoogle, v.GoogleID.String)
		}

		if v.PatreonID.Valid && v.PatreonID.String != "" {
			createUser(v.ID, oauth.ProviderPatreon, v.PatreonID.String)
		}

		if v.SteamID.Valid && v.SteamID.String != "" {
			createUser(v.ID, oauth.ProviderSteam, v.SteamID.String)
		}
	}
}

func createUser(userID int, enum oauth.ProviderEnum, providerID string) {

	db, err := mysql.GetMySQLClient()
	if err != nil {
		log.ErrS(err)
		return
	}

	userProvider := mysql.UserProvider{}
	userProvider.UserID = userID
	userProvider.Provider = enum
	userProvider.ID = providerID

	err = db.Save(&userProvider).Error
	if err != nil {
		log.ErrS(err)
	}
}
