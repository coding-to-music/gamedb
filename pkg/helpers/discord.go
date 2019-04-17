package helpers

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/website/pkg"
)

var (
	discordMutex   sync.Mutex
	discordSession *discordgo.Session
)

func GetDiscord(handlers ...interface{}) (session *discordgo.Session, err error) {

	discordMutex.Lock()
	defer discordMutex.Unlock()

	if discordSession == nil {

		discord, err := discordgo.New("Bot " + config.Config.DiscordRelayToken.Get())
		if err != nil {
			return discord, err
		}

		for _, v := range handlers {
			discord.AddHandler(v)
		}

		// Open connection
		err = discord.Open()
		if err != nil {
			return discord, err
		}

		discordSession = discord
	}

	return discordSession, err
}
