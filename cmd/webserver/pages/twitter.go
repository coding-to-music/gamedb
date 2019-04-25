package pages

import (
	"net/http"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/log"
	"github.com/go-chi/chi"
)

func TwitterRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", twitterHandler)
	return r
}

func twitterHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*1)

	t := twitterTemplate{}
	t.fill(w, r, "News", "")

	client := helpers.GetTwitter()

	tru := true
	tweets, _, err := client.Timelines.UserTimeline(&twitter.UserTimelineParams{
		ScreenName:     "gamedbonline",
		Count:          100,
		ExcludeReplies: &tru,
	})

	t.Tweets = tweets

	err = returnTemplate(w, r, "twitter", t)
	log.Err(err, r)
}

type twitterTemplate struct {
	GlobalTemplate
	Tweets []twitter.Tweet
}
