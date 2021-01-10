package tasks

import (
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/gocolly/colly/v2"
	"github.com/patrickmn/go-cache"
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

var (
	articlesLatestRegex = regexp.MustCompile(`/apps/([0-9]+)/`)
	articlesLatestCache = cache.New(time.Hour*24*7, time.Minute*30)
)

func (c ArticlesLatest) work() (err error) {

	col := colly.NewCollector(
		steam.WithTimeout(0),
	)

	// Tags
	col.OnHTML(`div[id^="post_"]`, func(e *colly.HTMLElement) {

		postID := e.Attr("id")

		if _, ok := articlesLatestCache.Get(postID); !ok {

			articlesLatestCache.SetDefault(postID, nil)

			sub := articlesLatestRegex.FindStringSubmatch(e.ChildAttr("img.capsule", "src"))
			if len(sub) == 2 {
				i, err := strconv.Atoi(sub[1])
				if err == nil {
					err = queue.ProduceAppNews(i)
					if err != nil {
						log.ErrS(err)
					}
				}
			}
		}
	})

	feeds, err := mongo.GetAppArticlesGroupedByFeed(0)
	if err != nil {
		return err
	}

	for _, v := range feeds {

		q := url.Values{}
		q.Set("feed", v.ID)
		q.Set("enddate", strconv.FormatInt(time.Now().Unix(), 10))

		err = col.Visit("https://store.steampowered.com/news/posts?" + q.Encode())
		if err != nil {
			return err
		}
	}

	return nil
}
