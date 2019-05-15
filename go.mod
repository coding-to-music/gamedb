module github.com/gamedb/gamedb

// These two lines are here because ahmdrz/goinsta needs to tag their repo after a fix
replace github.com/ahmdrz/goinsta/v2 => github.com/krylovsk/goinsta/v2 v2.4.0

require github.com/ahmdrz/goinsta/v2 v2.4.0

require (
	cloud.google.com/go v0.37.0
	github.com/Jleagle/go-durationfmt v0.0.0-20190307132420-e57bfad84057
	github.com/Jleagle/google-cloud-storage-go v0.0.0-20181227195340-0633133a5c6c
	github.com/Jleagle/influxql v0.0.0-20190502115937-4ac053a1ed8e
	github.com/Jleagle/memcache-go v0.0.0-20190306211229-e1529000ed89
	github.com/Jleagle/patreon-go v0.0.0-20190513114123-359f6ccef16d
	github.com/Jleagle/recaptcha-go v0.0.0-20190220085232-0e548dc7cc83
	github.com/Jleagle/session-go v0.0.0-20190514170753-50d0ff5e9e78
	github.com/Jleagle/sitemap-go v0.0.0-20190405195207-2bdddbb3bd50
	github.com/Jleagle/steam-go v0.0.0-20190515120823-3f2370388ffd
	github.com/PuerkitoBio/goquery v1.5.0 // indirect
	github.com/antchfx/htmlquery v1.0.0 // indirect
	github.com/antchfx/xmlquery v1.0.0 // indirect
	github.com/antchfx/xpath v0.0.0-20190129040759-c8489ed3251e // indirect
	github.com/badoux/checkmail v0.0.0-20181210160741-9661bd69e9ad
	github.com/bwmarrin/discordgo v0.19.0
	github.com/cenkalti/backoff v2.1.1+incompatible
	github.com/cockroachdb/apd v1.1.0 // indirect
	github.com/denisenkom/go-mssqldb v0.0.0-20190204142019-df6d76eb9289 // indirect
	github.com/derekstavis/go-qs v0.0.0-20180720192143-9eef69e6c4e7
	github.com/dghubble/go-twitter v0.0.0-20190108053744-7fd79e2bcc65
	github.com/dghubble/oauth1 v0.5.0
	github.com/dghubble/sling v1.2.0 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/frustra/bbcode v0.0.0-20180807171629-48be21ce690c
	github.com/go-chi/chi v4.0.1+incompatible
	github.com/go-chi/cors v1.0.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gocolly/colly v1.2.0
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/golang/protobuf v1.3.0
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/go-github v17.0.0+incompatible
	github.com/gorilla/sessions v1.1.3
	github.com/gorilla/websocket v1.4.0
	github.com/gosimple/slug v1.4.2
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/influxdata/influxdb1-client v0.0.0-20190124185755-16c852ea613f
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/jinzhu/now v1.0.0
	github.com/kennygrant/sanitize v1.2.4 // indirect
	github.com/leekchan/accounting v0.0.0-20180703100437-18a1925d6514
	github.com/lib/pq v1.0.0 // indirect
	github.com/logrusorgru/aurora v0.0.0-20181002194514-a7b3b318ed4e
	github.com/lusis/go-slackbot v0.0.0-20180109053408-401027ccfef5 // indirect
	github.com/lusis/slack-test v0.0.0-20190426140909-c40012f20018 // indirect
	github.com/mattn/go-sqlite3 v1.10.0 // indirect
	github.com/microcosm-cc/bluemonday v1.0.2
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mxpv/patreon-go v0.0.0-20180807002359-67dbab1ad14c
	github.com/nicklaw5/helix v0.5.1
	github.com/nlopes/slack v0.5.0
	github.com/pariz/gountries v0.0.0-20171019111738-adb00f6513a3
	github.com/pkg/errors v0.8.1 // indirect
	github.com/rainycape/unidecode v0.0.0-20150907023854-cb7f23ec59be // indirect
	github.com/robfig/cron v0.0.0-20180505203441-b41be1df6967
	github.com/russross/blackfriday v2.0.0+incompatible
	github.com/saintfish/chardet v0.0.0-20120816061221-3af4cd4741ca // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sendgrid/rest v2.4.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.4.1+incompatible
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24 // indirect
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	github.com/stretchr/testify v1.3.0 // indirect
	github.com/tdewolff/minify v2.3.6+incompatible
	github.com/tdewolff/parse v2.3.4+incompatible // indirect
	github.com/tdewolff/test v1.0.0 // indirect
	github.com/temoto/robotstxt v0.0.0-20180810133444-97ee4a9ee6ea // indirect
	github.com/tidwall/pretty v0.0.0-20180105212114-65a9db5fad51 // indirect
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	github.com/yohcop/openid-go v0.0.0-20170901155220-cfc72ed89575
	go.mongodb.org/mongo-driver v1.0.0
	go.opencensus.io v0.19.1 // indirect
	golang.org/x/crypto v0.0.0-20190404164418-38d8ce5564a5
	golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3 // indirect
	golang.org/x/oauth2 v0.0.0-20190226205417-e64efc72b421
	golang.org/x/sys v0.0.0-20190405154228-4b34438f7a67 // indirect
	google.golang.org/genproto v0.0.0-20190307195333-5fe7a883aa19
	google.golang.org/grpc v1.19.0
)
