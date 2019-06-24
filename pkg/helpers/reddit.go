package helpers

import (
	"log"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/jzelinskie/geddit"
)

var (
	redditSession *geddit.OAuthSession
	redditLock    sync.Mutex
)

func GetReddit() (session *geddit.OAuthSession, err error) {

	redditLock.Lock()
	defer redditLock.Unlock()

	if redditSession == nil {

		sess, err := geddit.NewOAuthSession("", "", "jzelinskie/geddit", "", )
		if err != nil {
			log.Fatal(err)
		}

		// Create new auth token for confidential clients (personal scripts/apps).
		err = sess.LoginAuth(config.Config.RedditUsername.Get(), config.Config.RedditPassword.Get())
		if err != nil {
			log.Fatal(err)
		}

		redditSession = sess
	}

	return redditSession, err
}
