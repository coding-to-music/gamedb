package helpers

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

var (
	discordConnections = map[string]*discordgo.Session{}
	discordMutex       sync.Mutex
)

func GetDiscordBot(authToken string, bot bool, handlers ...interface{}) (session *discordgo.Session, err error) {

	if bot {
		authToken = "Bot " + authToken
	} else {
		authToken = "Bearer " + authToken
	}

	discordMutex.Lock()
	defer discordMutex.Unlock()

	_, ok := discordConnections[authToken]
	if !ok {

		discord, err := discordgo.New(authToken)
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

		discordConnections[authToken] = discord
	}

	return discordConnections[authToken], err
}
