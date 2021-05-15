module github.com/gamedb/gamedb

go 1.13

replace github.com/getkin/kin-openapi v0.61.0 => github.com/getkin/kin-openapi v0.53.0

require (
	cloud.google.com/go/logging v1.4.1
	github.com/Jleagle/captcha-go v0.0.0-20201203202350-806e55b8099c
	github.com/Jleagle/go-durationfmt v0.0.0-20190307132420-e57bfad84057
	github.com/Jleagle/influxql v0.0.0-20200804190929-88324f67bffe
	github.com/Jleagle/memcache-go v0.0.0-20210309190441-636ebc04a889
	github.com/Jleagle/patreon-go v0.0.0-20201006180837-366bfaa6710a
	github.com/Jleagle/rabbit-go v0.0.0-20210115203259-266db76b636e
	github.com/Jleagle/rate-limit-go v0.0.0-20210514120325-52a5462241e1
	github.com/Jleagle/session-go v0.0.0-20190515070633-3c8712426233
	github.com/Jleagle/sitemap-go v0.0.0-20201217201247-75f3818f336a
	github.com/Jleagle/steam-go v0.0.0-20210211214415-6e5c1aecbba1
	github.com/Jleagle/unmarshal-go v0.0.0-20200217225147-fd7db71d9ac0
	github.com/Philipp15b/go-steam v1.0.1-0.20210301125207-f5f3f40fa791
	github.com/PuerkitoBio/goquery v1.6.1 // indirect
	github.com/StefanSchroeder/Golang-Roman v1.0.1-0.20210311185938-864df5cde20d
	github.com/ahmdrz/goinsta/v2 v2.4.5
	github.com/antchfx/xmlquery v1.3.6 // indirect
	github.com/antchfx/xpath v1.1.11 // indirect
	github.com/aws/aws-sdk-go v1.38.40 // indirect
	github.com/badoux/checkmail v1.2.1
	github.com/blend/go-sdk v1.1.1 // indirect
	github.com/bwmarrin/discordgo v0.23.3-0.20210314162722-182d9b48f34b
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/deepmap/oapi-codegen v1.6.1
	github.com/derekstavis/go-qs v0.0.0-20180720192143-9eef69e6c4e7
	github.com/dghubble/go-twitter v0.0.0-20201011215211-4b180d0cc78d
	github.com/dghubble/oauth1 v0.7.0
	github.com/digitalocean/godo v1.61.0
	github.com/djherbis/fscache v0.10.1
	github.com/dustin/go-humanize v1.0.0
	github.com/frustra/bbcode v0.0.0-20201127003707-6ef347fbe1c8
	github.com/fsnotify/fsnotify v1.4.9
	github.com/getkin/kin-openapi v0.61.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/go-chi/chi/v5 v5.0.3
	github.com/go-chi/cors v1.2.0
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gobuffalo/packr/v2 v2.8.1
	github.com/gocolly/colly/v2 v2.1.0
	github.com/golang/glog v0.0.0-20210429001901-424d2337a529 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/go-github/v32 v32.1.0
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.1
	github.com/gorilla/websocket v1.4.2
	github.com/gosimple/slug v1.9.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.4.0
	github.com/hetznercloud/hcloud-go v1.25.0
	github.com/influxdata/influxdb-client-go/v2 v2.3.0
	github.com/influxdata/influxdb1-client v0.0.0-20200827194710-b269163b24ab
	github.com/influxdata/line-protocol v0.0.0-20210311194329-9aa0e372d097 // indirect
	github.com/jinzhu/gorm v1.9.16
	github.com/jinzhu/now v1.1.2
	github.com/justinas/nosurf v1.1.1
	github.com/karrick/godirwalk v1.16.1 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/klauspost/compress v1.12.2 // indirect
	github.com/labstack/echo/v4 v4.3.0 // indirect
	github.com/lib/pq v1.3.0 // indirect
	github.com/mailjet/mailjet-apiv3-go v0.0.0-20201009050126-c24bc15a9394
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/mattn/go-sqlite3 v2.0.1+incompatible // indirect
	github.com/mborgerson/GoTruncateHtml v0.0.0-20150507032438-125d9154cd1e
	github.com/memcachier/mc/v3 v3.0.3 // indirect
	github.com/microcosm-cc/bluemonday v1.0.9
	github.com/montanaflynn/stats v0.6.6
	github.com/mssola/user_agent v0.5.2
	github.com/nicklaw5/helix v1.15.0
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/olivere/elastic/v7 v7.0.24
	github.com/oschwald/maxminddb-golang v1.8.0
	github.com/pariz/gountries v0.0.0-20200430155801-1c6a393df9c7
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/robfig/cron/v3 v3.0.1
	github.com/rogpeppe/go-internal v1.6.1 // indirect
	github.com/rollbar/rollbar-go v1.4.0
	github.com/russross/blackfriday v2.0.0+incompatible
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/streadway/amqp v1.0.0
	github.com/temoto/robotstxt v1.1.2 // indirect
	github.com/tidwall/pretty v1.0.1 // indirect
	github.com/wcharczuk/go-chart v2.0.1+incompatible
	github.com/yohcop/openid-go v1.0.0
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	go.mongodb.org/mongo-driver v1.5.2
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/exp v0.0.0-20200331195152-e8c3332aa8e5 // indirect
	golang.org/x/image v0.0.0-20210504121937-7319ad40d33e // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20210510120150-4163338589ed
	golang.org/x/oauth2 v0.0.0-20210514164344-f6687ab2804c
	golang.org/x/sys v0.0.0-20210514084401-e8d321eab015 // indirect
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56 // indirect
	golang.org/x/text v0.3.6
	golang.org/x/tools v0.1.1 // indirect
	gonum.org/v1/gonum v0.9.1
	google.golang.org/api v0.46.0
	google.golang.org/genproto v0.0.0-20210513213006-bf773b8c8384 // indirect
	google.golang.org/grpc v1.37.1
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/djherbis/atime.v1 v1.0.0 // indirect
	gopkg.in/djherbis/stream.v1 v1.3.1 // indirect
	gopkg.in/yaml.v2 v2.4.0
	jaytaylor.com/html2text v0.0.0-20200412013138-3577fbdbcff7
	mvdan.cc/xurls/v2 v2.2.0
)
