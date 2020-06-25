package reddit

import (
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/jzelinskie/geddit"
)

var (
	redditSession *geddit.LoginSession
	redditLock    sync.Mutex
)

func getReddit() (session *geddit.LoginSession, err error) {

	redditLock.Lock()
	defer redditLock.Unlock()

	if redditSession == nil {

		session, err = geddit.NewLoginSession(
			config.Config.RedditUsername.Get(),
			config.Config.RedditPassword.Get(),
			"jzelinskie/geddit",
		)

		if err != nil {
			return session, err
		}

		redditSession = session
	}

	return redditSession, err
}

func PostToReddit(title string, link string) (err error) {

	sess, err := getReddit()
	if err != nil {
		return err
	}

	return sess.Submit(geddit.NewLinkSubmission("gamedb", title, link, true, &geddit.Captcha{}))
}
