package helpers

import (
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/rollbar/rollbar-go"
)

func InitRollbar() {

	rollbar.SetToken(config.Config.RollbarSecret.Get())
	rollbar.SetEnvironment(config.Config.Environment.Get())
	rollbar.SetCodeVersion(config.Config.CommitHash.Get())
	rollbar.SetServerHost("gamedb.online")
	rollbar.SetServerRoot("github.com/gamedb/gamedb")
}
