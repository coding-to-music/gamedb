package helpers

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

var (
	discordConnections = map[string]*discordgo.Session{}
	discordMutex       sync.Mutex
)

func GetDiscordBot(botToken string, handlers ...interface{}) (session *discordgo.Session, err error) {

	discordMutex.Lock()
	defer discordMutex.Unlock()

	_, ok := discordConnections[botToken]
	if !ok {

		discord, err := discordgo.New("Bot " + botToken)
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

		discordConnections[botToken] = discord
	}

	return discordConnections[botToken], err
}
