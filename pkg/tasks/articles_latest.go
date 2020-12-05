package tasks

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/queue"
)

type ArticlesLatest struct {
	BaseTask
}

func (c ArticlesLatest) ID() string {
	return "news-latest"
}

func (c ArticlesLatest) Name() string {
	return "Grab the latest news articles from steam"
}

func (c ArticlesLatest) Group() TaskGroup {
	return TaskGroupNews
}

func (c ArticlesLatest) Cron() TaskTime {
	return CronTimeNewsLatest
}

var articlesLatest = regexp.MustCompile(`/apps/([0-9]+)/`)

func (c ArticlesLatest) work() (err error) {

	q := url.Values{}
	q.Set("feed", "steam_community_announcements")
	q.Set("enddate", strconv.FormatInt(time.Now().Unix(), 10))

	b, code, err := helpers.Get("https://store.steampowered.com/news/posts?"+q.Encode(), 0, nil)
	if err != nil {
		return err
	}
	if code != 200 {
		return errors.New("bad response")
	}

	for _, v := range articlesLatest.FindAllStringSubmatch(string(b), -1) {
		if len(v) == 2 {
			i, err := strconv.Atoi(v[1])
			if err == nil {
				err = queue.ProduceAppNews(i)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
