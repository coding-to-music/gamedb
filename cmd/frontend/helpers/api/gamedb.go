package api

import (
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/getkin/kin-openapi/openapi3"
)

const (
	tagGames    = "Games"
	tagPlayers  = "Players"
	tagArticles = "Articles"
	tagPackages = "Packages"
	tagGroups   = "Groups"
)

func stringPointer(s string) *string {
	return &s
}

var SwaggerGameDB = &openapi3.Swagger{
	OpenAPI: "3.0.0",
	Servers: []*openapi3.Server{
		{URL: "https://api.gamedb.online"},
	},
	ExternalDocs: &openapi3.ExternalDocs{
		URL: config.C.GameDBDomain + "/api/gamedb",
	},
	Info: &openapi3.Info{
		Title:          "Game DB API",
		Version:        "1.0.0",
		TermsOfService: config.C.GameDBDomain + "/terms",
		Contact: &openapi3.Contact{
			Name: "Jleagle",
			URL:  config.C.GameDBDomain + "/contact",
		},
		ExtensionProps: openapi3.ExtensionProps{Extensions: map[string]interface{}{
			"x-logo": config.C.GameDBDomain + "/assets/img/sa-bg-192x192.png",
		}},
	},
	Tags: openapi3.Tags{
		&openapi3.Tag{Name: tagGames},
		&openapi3.Tag{Name: tagPlayers},
		&openapi3.Tag{Name: tagArticles},
		&openapi3.Tag{Name: tagPackages},
		&openapi3.Tag{Name: tagGroups},
	},
	Security: openapi3.SecurityRequirements{
		openapi3.NewSecurityRequirement().Authenticate("key-header"),
		openapi3.NewSecurityRequirement().Authenticate("key-query"),
	},
	Components: openapi3.Components{
		SecuritySchemes: map[string]*openapi3.SecuritySchemeRef{
			"key-header": {Value: openapi3.NewSecurityScheme().WithName("key").WithType("apiKey").WithIn("header")},
			"key-query":  {Value: openapi3.NewSecurityScheme().WithName("key").WithType("apiKey").WithIn("query")},
		},
		Parameters: map[string]*openapi3.ParameterRef{
			"limit-param": {
				Value: openapi3.NewQueryParameter("limit").WithSchema(openapi3.NewIntegerSchema().WithDefault(10).WithMin(1).WithMax(1000)),
			},
			"offset-param": {
				Value: openapi3.NewQueryParameter("offset").WithSchema(openapi3.NewIntegerSchema().WithDefault(0).WithMin(0)),
			},
			"order-param-asc": {
				Value: openapi3.NewQueryParameter("order").WithSchema(openapi3.NewStringSchema().WithEnum("asc", "desc").WithDefault("asc")),
			},
			"order-param-desc": {
				Value: openapi3.NewQueryParameter("order").WithSchema(openapi3.NewStringSchema().WithEnum("asc", "desc").WithDefault("desc")),
			},
		},
		Schemas: map[string]*openapi3.SchemaRef{
			"pagination-schema": {
				Value: &openapi3.Schema{
					Required: []string{"offset", "limit", "total", "pagesTotal", "pagesCurrent"},
					Properties: map[string]*openapi3.SchemaRef{
						"offset":       {Value: openapi3.NewInt64Schema()},
						"limit":        {Value: openapi3.NewInt64Schema()},
						"total":        {Value: openapi3.NewInt64Schema()},
						"pagesTotal":   {Value: openapi3.NewInt64Schema()},
						"pagesCurrent": {Value: openapi3.NewInt64Schema()},
					},
				},
			},
			"app-schema": {
				Value: &openapi3.Schema{
					Required: []string{"id", "name", "tags", "genres", "categories", "developers", "publishers", "prices", "players_max", "players_week_max", "players_week_avg", "release_date", "reviews_positive", "reviews_negative", "reviews_score", "metacritic_score"},
					Properties: map[string]*openapi3.SchemaRef{
						"id":               {Value: openapi3.NewIntegerSchema()},
						"name":             {Value: openapi3.NewStringSchema()},
						"tags":             {Value: &openapi3.Schema{Type: "array", Items: &openapi3.SchemaRef{Ref: "#/components/schemas/stat-schema"}}},
						"genres":           {Value: &openapi3.Schema{Type: "array", Items: &openapi3.SchemaRef{Ref: "#/components/schemas/stat-schema"}}},
						"categories":       {Value: &openapi3.Schema{Type: "array", Items: &openapi3.SchemaRef{Ref: "#/components/schemas/stat-schema"}}},
						"developers":       {Value: &openapi3.Schema{Type: "array", Items: &openapi3.SchemaRef{Ref: "#/components/schemas/stat-schema"}}},
						"publishers":       {Value: &openapi3.Schema{Type: "array", Items: &openapi3.SchemaRef{Ref: "#/components/schemas/stat-schema"}}},
						"prices":           {Value: &openapi3.Schema{Type: "object", AdditionalProperties: &openapi3.SchemaRef{Ref: "#/components/schemas/product-price-schema"}}},
						"players_max":      {Value: openapi3.NewIntegerSchema()},
						"players_week_max": {Value: openapi3.NewIntegerSchema()},
						"release_date":     {Value: openapi3.NewInt64Schema()},
						"reviews_positive": {Value: openapi3.NewIntegerSchema()},
						"reviews_negative": {Value: openapi3.NewIntegerSchema()},
						"reviews_score":    {Value: openapi3.NewFloat64Schema().WithFormat("double")},
						"metacritic_score": {Value: openapi3.NewInt32Schema()},
						// "players_week_avg": {Value: openapi3.NewFloat64Schema().WithFormat("double")},
					},
				},
			},
			"article-schema": {
				Value: &openapi3.Schema{
					Required: []string{"id", "title", "url", "author", "contents", "date", "feed_label", "feed", "feed_type", "app_id", "app_icon", "icon"},
					Properties: map[string]*openapi3.SchemaRef{
						"id":         {Value: openapi3.NewInt64Schema()},
						"title":      {Value: openapi3.NewStringSchema()},
						"url":        {Value: openapi3.NewStringSchema()},
						"author":     {Value: openapi3.NewStringSchema()},
						"contents":   {Value: openapi3.NewStringSchema()},
						"date":       {Value: openapi3.NewInt64Schema()},
						"feed_label": {Value: openapi3.NewStringSchema()},
						"feed":       {Value: openapi3.NewStringSchema()},
						"feed_type":  {Value: openapi3.NewInt32Schema()},
						"app_id":     {Value: openapi3.NewInt32Schema()},
						"app_icon":   {Value: openapi3.NewStringSchema()},
						"icon":       {Value: openapi3.NewStringSchema()},
					},
				},
			},
			"group-schema": {
				Value: &openapi3.Schema{
					Required: []string{"id", "name", "abbreviation", "url", "app_id", "headline", "icon", "trending", "members", "members_in_chat", "members_in_game", "members_online", "error", "type", "primaries"},
					Properties: map[string]*openapi3.SchemaRef{
						"id":              {Value: openapi3.NewStringSchema()},
						"name":            {Value: openapi3.NewStringSchema()},
						"abbreviation":    {Value: openapi3.NewStringSchema()},
						"url":             {Value: openapi3.NewStringSchema()},
						"app_id":          {Value: openapi3.NewInt32Schema()},
						"headline":        {Value: openapi3.NewStringSchema()},
						"icon":            {Value: openapi3.NewStringSchema()},
						"trending":        {Value: openapi3.NewFloat64Schema()},
						"members":         {Value: openapi3.NewInt32Schema()},
						"members_in_chat": {Value: openapi3.NewInt32Schema()},
						"members_in_game": {Value: openapi3.NewInt32Schema()},
						"members_online":  {Value: openapi3.NewInt32Schema()},
						"error":           {Value: openapi3.NewStringSchema()},
						"type":            {Value: openapi3.NewStringSchema()},
						"primaries":       {Value: openapi3.NewInt32Schema()},
					},
				},
			},
			"package-schema": {
				Value: &openapi3.Schema{
					Required: []string{"apps", "apps_count", "bundle", "billing_type", "change_id", "change_number_date", "coming_soon", "depot_ids", "icon", "id", "image_logo", "image_page", "license_type", "name", "platforms", "prices", "release_date", "release_date_unix", "status"},
					Properties: map[string]*openapi3.SchemaRef{
						"apps":               {Value: openapi3.NewArraySchema().WithItems(openapi3.NewInt32Schema())},
						"apps_count":         {Value: openapi3.NewInt32Schema()},
						"bundle":             {Value: openapi3.NewArraySchema().WithItems(openapi3.NewInt32Schema())},
						"billing_type":       {Value: openapi3.NewStringSchema()},
						"change_id":          {Value: openapi3.NewInt32Schema()},
						"change_number_date": {Value: openapi3.NewInt64Schema()},
						"coming_soon":        {Value: openapi3.NewBoolSchema()},
						"depot_ids":          {Value: openapi3.NewArraySchema().WithItems(openapi3.NewInt32Schema())},
						"icon":               {Value: openapi3.NewStringSchema()},
						"id":                 {Value: openapi3.NewInt32Schema()},
						"image_logo":         {Value: openapi3.NewStringSchema()},
						"image_page":         {Value: openapi3.NewStringSchema()},
						"license_type":       {Value: openapi3.NewStringSchema()},
						"name":               {Value: openapi3.NewStringSchema()},
						"platforms":          {Value: openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())},
						"prices":             {Value: &openapi3.Schema{Type: "object", AdditionalProperties: &openapi3.SchemaRef{Ref: "#/components/schemas/product-price-schema"}}},
						"release_date":       {Value: openapi3.NewStringSchema()},
						"release_date_unix":  {Value: openapi3.NewInt64Schema()},
						"status":             {Value: openapi3.NewStringSchema()},
					},
				},
			},
			"product-price-schema": {
				Value: &openapi3.Schema{
					Required: []string{"currency", "initial", "final", "discountPercent", "individual", "free"},
					Properties: map[string]*openapi3.SchemaRef{
						"currency":        {Value: openapi3.NewStringSchema()},
						"initial":         {Value: openapi3.NewInt32Schema()},
						"final":           {Value: openapi3.NewInt32Schema()},
						"discountPercent": {Value: openapi3.NewInt32Schema()},
						"individual":      {Value: openapi3.NewInt32Schema()},
						"free":            {Value: openapi3.NewBoolSchema()},
					},
				},
			},
			"stat-schema": {
				Value: &openapi3.Schema{
					Required: []string{"id", "name"},
					Properties: map[string]*openapi3.SchemaRef{
						"id":   {Value: openapi3.NewIntegerSchema()},
						"name": {Value: openapi3.NewStringSchema()},
					},
				},
			},
			"player-schema": {
				Value: &openapi3.Schema{
					Required: []string{"id", "name", "avatar", "badges", "comments", "friends", "games", "groups", "level", "playtime", "country", "continent", "state", "vanity_url"},
					Properties: map[string]*openapi3.SchemaRef{
						"id":         {Value: openapi3.NewStringSchema()}, // Too big for int in JS
						"name":       {Value: openapi3.NewStringSchema()},
						"avatar":     {Value: openapi3.NewStringSchema()},
						"badges":     {Value: openapi3.NewIntegerSchema()},
						"comments":   {Value: openapi3.NewIntegerSchema()},
						"friends":    {Value: openapi3.NewIntegerSchema()},
						"games":      {Value: openapi3.NewIntegerSchema()},
						"groups":     {Value: openapi3.NewIntegerSchema()},
						"level":      {Value: openapi3.NewIntegerSchema()},
						"playtime":   {Value: openapi3.NewIntegerSchema()},
						"country":    {Value: openapi3.NewStringSchema()},
						"continent":  {Value: openapi3.NewStringSchema()},
						"state":      {Value: openapi3.NewStringSchema()},
						"vanity_url": {Value: openapi3.NewStringSchema()},
					},
				},
			},
			"message-schema": {
				Value: &openapi3.Schema{
					Required: []string{"message"},
					Properties: map[string]*openapi3.SchemaRef{
						"message": {Value: openapi3.NewStringSchema()},
					},
				},
			},
			"price-schema": {
				Ref: "#/components/schemas/product-price-schema",
			},
		},
		Responses: map[string]*openapi3.ResponseRef{
			"message-response": {
				Value: &openapi3.Response{
					ExtensionProps: openapi3.ExtensionProps{},
					Description:    stringPointer("Message"),
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/message-schema",
					}),
				},
			},
			"pagination-response": {
				Value: &openapi3.Response{
					Description: stringPointer("Page information"),
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/pagination-schema",
					}),
				},
			},
			"article-response": {
				Value: &openapi3.Response{
					Description: stringPointer("A article"),
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/article-schema",
					}),
				},
			},
			"articles-response": {
				Value: &openapi3.Response{
					Description: stringPointer("List of articles"),
					Content: openapi3.NewContentWithJSONSchema(&openapi3.Schema{
						Description: "List of articles",
						Required:    []string{"pagination", "articles"},
						Properties: map[string]*openapi3.SchemaRef{
							"pagination": {
								Ref: "#/components/schemas/pagination-schema",
							},
							"articles": {
								Value: &openapi3.Schema{
									Type: "array",
									Items: &openapi3.SchemaRef{
										Ref: "#/components/schemas/article-schema",
									},
								},
							},
						},
					}),
				},
			},
			"app-response": {
				Value: &openapi3.Response{
					Description: stringPointer("A game"),
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/app-schema",
					}),
				},
			},
			"apps-response": {
				Value: &openapi3.Response{
					Description: stringPointer("List of games"),
					Content: openapi3.NewContentWithJSONSchema(&openapi3.Schema{
						Description: "List of apps, with pagination",
						Required:    []string{"pagination", "apps"},
						Properties: map[string]*openapi3.SchemaRef{
							"pagination": {
								Ref: "#/components/schemas/pagination-schema",
							},
							"apps": {
								Value: &openapi3.Schema{
									Type: "array",
									Items: &openapi3.SchemaRef{
										Ref: "#/components/schemas/app-schema",
									},
								},
							},
						},
					}),
				},
			},
			"group-response": {
				Value: &openapi3.Response{
					Description: stringPointer("A group"),
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/group-schema",
					}),
				},
			},
			"groups-response": {
				Value: &openapi3.Response{
					Description: stringPointer("List of groups"),
					Content: openapi3.NewContentWithJSONSchema(&openapi3.Schema{
						Description: "List of groups",
						Required:    []string{"pagination", "groups"},
						Properties: map[string]*openapi3.SchemaRef{
							"pagination": {
								Ref: "#/components/schemas/pagination-schema",
							},
							"groups": {
								Value: &openapi3.Schema{
									Type: "array",
									Items: &openapi3.SchemaRef{
										Ref: "#/components/schemas/group-schema",
									},
								},
							},
						},
					}),
				},
			},
			"package-response": {
				Value: &openapi3.Response{
					Description: stringPointer("A package"),
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/package-schema",
					}),
				},
			},
			"packages-response": {
				Value: &openapi3.Response{
					Description: stringPointer("List of games"),
					Content: openapi3.NewContentWithJSONSchema(&openapi3.Schema{
						Description: "List of packages",
						Required:    []string{"pagination", "packages"},
						Properties: map[string]*openapi3.SchemaRef{
							"pagination": {
								Ref: "#/components/schemas/pagination-schema",
							},
							"packages": {
								Value: &openapi3.Schema{
									Type: "array",
									Items: &openapi3.SchemaRef{
										Ref: "#/components/schemas/package-schema",
									},
								},
							},
						},
					}),
				},
			},
			"player-response": {
				Value: &openapi3.Response{
					Description: stringPointer("A player"),
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/player-schema",
					}),
				},
			},
			"players-response": {
				Value: &openapi3.Response{
					Description: stringPointer("List of players"),
					Content: openapi3.NewContentWithJSONSchema(&openapi3.Schema{
						Required: []string{"pagination", "players"},
						Properties: map[string]*openapi3.SchemaRef{
							"pagination": {
								Ref: "#/components/schemas/pagination-schema",
							},
							"players": {
								Value: &openapi3.Schema{
									Type: "array",
									Items: &openapi3.SchemaRef{
										Ref: "#/components/schemas/player-schema",
									},
								},
							},
						},
					}),
				},
			},
		},
	},
	Paths: openapi3.Paths{
		"/articles": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Tags:    []string{tagArticles},
				Summary: "List Articles",
				Parameters: openapi3.Parameters{
					{Ref: "#/components/parameters/offset-param"},
					{Ref: "#/components/parameters/limit-param"},
					{Ref: "#/components/parameters/order-param-desc"},
					{Value: openapi3.NewQueryParameter("sort").WithSchema(openapi3.NewStringSchema())},
					{Value: openapi3.NewQueryParameter("ids").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewInt32Schema()).WithMaxItems(100))},
					{Value: openapi3.NewQueryParameter("app_ids").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewInt32Schema()).WithMaxItems(100))},
					{Value: openapi3.NewQueryParameter("feed").WithSchema(openapi3.NewStringSchema())},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/articles-response",
					},
					"401": {
						Ref: "#/components/responses/message-response",
					},
					"500": {
						Ref: "#/components/responses/message-response",
					},
				},
			},
		},
		"/articles/{id}": &openapi3.PathItem{},
		"/games": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Tags:    []string{tagGames},
				Summary: "List Games",
				Parameters: openapi3.Parameters{
					{Ref: "#/components/parameters/offset-param"},
					{Ref: "#/components/parameters/limit-param"},
					{Ref: "#/components/parameters/order-param-desc"},
					{Value: openapi3.NewQueryParameter("sort").WithSchema(openapi3.NewStringSchema())},
					{Value: openapi3.NewQueryParameter("ids").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewInt32Schema()).WithMaxItems(100))},
					{Value: openapi3.NewQueryParameter("tags").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewInt32Schema()).WithMaxItems(10))},
					{Value: openapi3.NewQueryParameter("genres").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewInt32Schema()).WithMaxItems(10))},
					{Value: openapi3.NewQueryParameter("categories").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewInt32Schema()).WithMaxItems(10))},
					{Value: openapi3.NewQueryParameter("developers").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewInt32Schema()).WithMaxItems(10))},
					{Value: openapi3.NewQueryParameter("publishers").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewInt32Schema()).WithMaxItems(10))},
					{Value: openapi3.NewQueryParameter("platforms").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema()).WithMaxItems(3))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/apps-response",
					},
					"401": {
						Ref: "#/components/responses/message-response",
					},
					"500": {
						Ref: "#/components/responses/message-response",
					},
				},
			},
		},
		"/games/{id}": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Tags:    []string{tagGames},
				Summary: "Retrieve Game",
				Parameters: openapi3.Parameters{
					{Value: openapi3.NewPathParameter("id").WithRequired(true).WithSchema(openapi3.NewInt32Schema().WithMin(1))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/app-response",
					},
					"400": {
						Ref: "#/components/responses/message-response",
					},
					"401": {
						Ref: "#/components/responses/message-response",
					},
					"404": {
						Ref: "#/components/responses/message-response",
					},
					"500": {
						Ref: "#/components/responses/message-response",
					},
				},
			},
		},
		"/groups": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Tags:    []string{tagGroups},
				Summary: "List Groups",
				Parameters: openapi3.Parameters{
					{Ref: "#/components/parameters/offset-param"},
					{Ref: "#/components/parameters/limit-param"},
					{Ref: "#/components/parameters/order-param-desc"},
					{Value: openapi3.NewQueryParameter("sort").WithSchema(openapi3.NewStringSchema().WithEnum("id", "members", "trending", "primaries").WithDefault("id"))},
					{Value: openapi3.NewQueryParameter("ids").WithSchema(openapi3.NewArraySchema().WithMaxItems(10).WithItems(openapi3.NewInt64Schema()))},
					{Value: openapi3.NewQueryParameter("type").WithSchema(openapi3.NewArraySchema().WithMaxItems(2).WithItems(openapi3.NewStringSchema()))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/groups-response",
					},
					"400": {
						Ref: "#/components/responses/message-response",
					},
					"401": {
						Ref: "#/components/responses/message-response",
					},
					"404": {
						Ref: "#/components/responses/message-response",
					},
					"500": {
						Ref: "#/components/responses/message-response",
					},
				},
			},
		},
		"/groups/{id}": &openapi3.PathItem{},
		"/packages": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Tags:    []string{tagPackages},
				Summary: "List Packages",
				Parameters: openapi3.Parameters{
					{Ref: "#/components/parameters/offset-param"},
					{Ref: "#/components/parameters/limit-param"},
					{Ref: "#/components/parameters/order-param-desc"},
					{Value: openapi3.NewQueryParameter("sort").WithSchema(openapi3.NewStringSchema().WithEnum("id").WithDefault("id"))},
					{Value: openapi3.NewQueryParameter("ids").WithSchema(openapi3.NewArraySchema().WithMaxItems(10).WithItems(openapi3.NewInt32Schema()))},
					{Value: openapi3.NewQueryParameter("billingType").WithSchema(openapi3.NewArraySchema().WithMaxItems(10).WithItems(openapi3.NewInt32Schema()))},
					{Value: openapi3.NewQueryParameter("licenseType").WithSchema(openapi3.NewArraySchema().WithMaxItems(10).WithItems(openapi3.NewInt32Schema()))},
					{Value: openapi3.NewQueryParameter("status").WithSchema(openapi3.NewArraySchema().WithMaxItems(10).WithItems(openapi3.NewInt32Schema()))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/packages-response",
					},
					"400": {
						Ref: "#/components/responses/message-response",
					},
					"401": {
						Ref: "#/components/responses/message-response",
					},
					"404": {
						Ref: "#/components/responses/message-response",
					},
					"500": {
						Ref: "#/components/responses/message-response",
					},
				},
			},
		},
		"/packages/{id}": &openapi3.PathItem{},
		"/players": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Tags:    []string{tagPlayers},
				Summary: "List Players",
				Parameters: openapi3.Parameters{
					{Ref: "#/components/parameters/offset-param"},
					{Ref: "#/components/parameters/limit-param"},
					{Ref: "#/components/parameters/order-param-desc"},
					{Value: openapi3.NewQueryParameter("sort").WithSchema(openapi3.NewStringSchema().WithEnum("id", "level", "badges", "games", "time", "friends", "comments").WithDefault("id"))},
					{Value: openapi3.NewQueryParameter("continent").WithSchema(openapi3.NewArraySchema().WithMaxItems(3).WithItems(openapi3.NewStringSchema().WithMaxLength(2)))},
					{Value: openapi3.NewQueryParameter("country").WithSchema(openapi3.NewArraySchema().WithMaxItems(3).WithItems(openapi3.NewStringSchema().WithMaxLength(2)))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/players-response",
					},
					"400": {
						Ref: "#/components/responses/message-response",
					},
					"401": {
						Ref: "#/components/responses/message-response",
					},
					"404": {
						Ref: "#/components/responses/message-response",
					},
					"500": {
						Ref: "#/components/responses/message-response",
					},
				},
			},
		},
		"/players/{id}": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Tags:    []string{tagPlayers},
				Summary: "Retrieve Player",
				Parameters: openapi3.Parameters{
					{Value: openapi3.NewPathParameter("id").WithRequired(true).WithSchema(openapi3.NewInt64Schema().WithMin(1))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/player-response",
					},
				},
			},
			Post: &openapi3.Operation{
				Tags:    []string{tagPlayers},
				Summary: "Update Player",
				Parameters: openapi3.Parameters{
					{Value: openapi3.NewPathParameter("id").WithRequired(true).WithSchema(openapi3.NewInt64Schema().WithMaxLength(2))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/message-response",
					},
					"401": {
						Ref: "#/components/responses/message-response",
					},
					"500": {
						Ref: "#/components/responses/message-response",
					},
				},
			},
		},
		// "/app - players",
		// "/app - price changes",
		// "/bundles",
		// "/bundles",
		// "/bundles/{id}",
		// "/changes",
		// "/players/{id}/update"
		// "/players/{id}/badges"
		// "/players/{id}/games"
		// "/players/{id}/history"
		// "/stats/Categories"
		// "/stats/Genres"
		// "/stats/Publishers"
		// "/stats/Steam"
		// "/stats/Tags"
	},
}
