module github.com/gamedb/website

// These two lines are here because ahmdrz/goinsta needs to tag their repo after a fix
replace github.com/ahmdrz/goinsta/v2 => github.com/krylovsk/goinsta/v2 v2.4.0

require github.com/ahmdrz/goinsta/v2 v2.4.0

require (
	cloud.google.com/go v0.36.0
	github.com/99designs/basicauth-go v0.0.0-20160802081356-2a93ba0f464d
	github.com/Jleagle/go-durationfmt v0.0.0-20190212102610-5ae8bf56bcbe
	github.com/Jleagle/google-cloud-storage-go v0.0.0-20181227195340-0633133a5c6c
	github.com/Jleagle/influxql v0.0.0-20190303215204-7c739e39d8a6
	github.com/Jleagle/memcache-go v0.0.0-20190306211229-e1529000ed89
	github.com/Jleagle/rabbit-go v0.0.0-20190220085424-6afd4589ce23
	github.com/Jleagle/recaptcha-go v0.0.0-20190220085232-0e548dc7cc83
	github.com/Jleagle/steam-go v0.0.0-20190220084322-c18ad4f60799
	github.com/PuerkitoBio/goquery v1.5.0 // indirect
	github.com/antchfx/htmlquery v1.0.0 // indirect
	github.com/antchfx/xmlquery v1.0.0 // indirect
	github.com/antchfx/xpath v0.0.0-20190129040759-c8489ed3251e // indirect
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
	github.com/google/go-github v17.0.0+incompatible
	github.com/gorilla/sessions v1.1.3
	github.com/gorilla/websocket v1.4.0
	github.com/gosimple/slug v1.4.2
	github.com/influxdata/influxdb1-client v0.0.0-20190124185755-16c852ea613f
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/jinzhu/now v1.0.0
	github.com/kennygrant/sanitize v1.2.4 // indirect
	github.com/leekchan/accounting v0.0.0-20180703100437-18a1925d6514
	github.com/lib/pq v1.0.0 // indirect
	github.com/logrusorgru/aurora v0.0.0-20181002194514-a7b3b318ed4e
	github.com/mattn/go-sqlite3 v1.10.0 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pariz/gountries v0.0.0-20171019111738-adb00f6513a3
	github.com/pkg/errors v0.8.1 // indirect
	github.com/rainycape/unidecode v0.0.0-20150907023854-cb7f23ec59be // indirect
	github.com/robfig/cron v0.0.0-20180505203441-b41be1df6967
	github.com/saintfish/chardet v0.0.0-20120816061221-3af4cd4741ca // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sendgrid/rest v2.4.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.4.1+incompatible
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24 // indirect
	github.com/streadway/amqp v0.0.0-20190214183023-884228600bc9
	github.com/tdewolff/minify v2.3.6+incompatible
	github.com/tdewolff/parse v2.3.4+incompatible // indirect
	github.com/tdewolff/test v1.0.0 // indirect
	github.com/temoto/robotstxt v0.0.0-20180810133444-97ee4a9ee6ea // indirect
	github.com/yohcop/openid-go v0.0.0-20170901155220-cfc72ed89575
	golang.org/x/crypto v0.0.0-20190219172222-a4c6cb3142f2
	golang.org/x/oauth2 v0.0.0-20190219183015-4b83411ed2b3
)
