module github.com/gamedb/gamedb

go 1.13

require (
	cloud.google.com/go v0.55.0 // indirect
	cloud.google.com/go/logging v1.0.0
	cloud.google.com/go/pubsub v1.3.1
	github.com/Jleagle/go-durationfmt v0.0.0-20190307132420-e57bfad84057
	github.com/Jleagle/influxql v0.0.0-20190502115937-4ac053a1ed8e
	github.com/Jleagle/memcache-go v0.0.0-20191228144235-986fe282434d
	github.com/Jleagle/patreon-go v0.0.0-20200117215733-2b8b00d4eab0
	github.com/Jleagle/rabbit-go v0.0.0-20200313125543-a2de27528286
	github.com/Jleagle/recaptcha-go v0.0.0-20200117124940-d00b2c62c076
	github.com/Jleagle/session-go v0.0.0-20190515070633-3c8712426233
	github.com/Jleagle/sitemap-go v0.0.0-20190405195207-2bdddbb3bd50
	github.com/Jleagle/steam-go v0.0.0-20200309212843-073d484199fa
	github.com/Jleagle/unmarshal-go v0.0.0-20200217225147-fd7db71d9ac0
	github.com/Philipp15b/go-steam v1.0.1-0.20190816133340-b04c5a83c1c0
	github.com/PuerkitoBio/goquery v1.5.1 // indirect
	github.com/ahmdrz/goinsta/v2 v2.4.5
	github.com/antchfx/htmlquery v1.2.2 // indirect
	github.com/antchfx/xmlquery v1.2.3 // indirect
	github.com/antchfx/xpath v1.1.4 // indirect
	github.com/badoux/checkmail v0.0.0-20181210160741-9661bd69e9ad
	github.com/beefsack/go-rate v0.0.0-20180408011153-efa7637bb9b6 // indirect
	github.com/buger/jsonparser v0.0.0-20200322175846-f7e751efca13 // indirect
	github.com/bwmarrin/discordgo v0.20.2
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.0.0
	github.com/deepmap/oapi-codegen v1.3.7-0.20200306152314-a80789f547c1
	github.com/derekstavis/go-qs v0.0.0-20180720192143-9eef69e6c4e7
	github.com/dghubble/go-twitter v0.0.0-20190719072343-39e5462e111f
	github.com/dghubble/oauth1 v0.6.0
	github.com/didip/tollbooth v4.0.2+incompatible
	github.com/didip/tollbooth/v5 v5.1.0
	github.com/djherbis/fscache v0.10.0
	github.com/dustin/go-humanize v1.0.0
	github.com/frustra/bbcode v0.0.0-20180807171629-48be21ce690c
	github.com/getkin/kin-openapi v0.3.0
	github.com/getsentry/sentry-go v0.5.1
	github.com/go-chi/chi v4.0.4+incompatible
	github.com/go-chi/cors v1.0.1
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gobuffalo/packr/v2 v2.8.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gocolly/colly v1.2.0
	github.com/golang/protobuf v1.3.5
	github.com/google/go-github/v28 v28.1.1
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.0
	github.com/gorilla/websocket v1.4.2
	github.com/gosimple/slug v1.9.0
	github.com/influxdata/influxdb1-client v0.0.0-20191209144304-8bf82d3c094d
	github.com/jinzhu/gorm v1.9.12
	github.com/jinzhu/now v1.1.1
	github.com/justinas/nosurf v1.1.0
	github.com/jzelinskie/geddit v0.0.0-20190913104144-95ef6806b073
	github.com/karrick/godirwalk v1.15.5 // indirect
	github.com/kennygrant/sanitize v1.2.4 // indirect
	github.com/klauspost/compress v1.10.3 // indirect
	github.com/labstack/echo/v4 v4.1.15 // indirect
	github.com/logrusorgru/aurora v0.0.0-20200102142835-e9ef32dff381
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/memcachier/mc v2.0.1+incompatible // indirect
	github.com/microcosm-cc/bluemonday v1.0.2
	github.com/mxpv/patreon-go v0.0.0-20190917022727-646111f1d983
	github.com/nicklaw5/helix v0.5.8
	github.com/nlopes/slack v0.6.0
	github.com/olekukonko/tablewriter v0.0.4 // indirect
	github.com/oschwald/maxminddb-golang v1.6.0
	github.com/pariz/gountries v0.0.0-20191029140926-233bc78cf5b5
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/powerslacker/ratelimit v0.0.0-20190505003410-df2fcffc8e0d
	github.com/robfig/cron/v3 v3.0.1
	github.com/rollbar/rollbar-go v1.2.0
	github.com/russross/blackfriday v2.0.0+incompatible
	github.com/saintfish/chardet v0.0.0-20120816061221-3af4cd4741ca // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sendgrid/rest v2.4.1+incompatible
	github.com/sendgrid/sendgrid-go v3.5.0+incompatible
	github.com/sirupsen/logrus v1.5.0 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/tdewolff/minify/v2 v2.7.3
	github.com/temoto/robotstxt v1.1.1 // indirect
	github.com/tidwall/pretty v1.0.1
	github.com/uber-go/atomic v1.4.0 // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	github.com/yohcop/openid-go v1.0.0
	go.mongodb.org/mongo-driver v1.3.1
	golang.org/x/crypto v0.0.0-20200323165209-0ec3e9974c59
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/text v0.3.2
	golang.org/x/tools v0.0.0-20200325010219-a49f79bcc224 // indirect
	google.golang.org/api v0.20.0
	google.golang.org/genproto v0.0.0-20200325114520-5b2d0af7952b // indirect
	google.golang.org/grpc v1.28.0
	gopkg.in/djherbis/atime.v1 v1.0.0 // indirect
	gopkg.in/djherbis/stream.v1 v1.3.0 // indirect
	jaytaylor.com/html2text v0.0.0-20200220170450-61d9dc4d7195
)
