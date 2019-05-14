package pages

import (
	"net/http"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
)

func TwitterRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", twitterHandler)
	return r
}

func twitterHandler(w http.ResponseWriter, r *http.Request) {

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
