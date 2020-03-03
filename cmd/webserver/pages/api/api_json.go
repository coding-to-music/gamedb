package api

import (
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/getkin/kin-openapi/openapi3"
)

func init() {

	err := openapi3.NewSwaggerLoader().ResolveRefsIn(Swagger, nil)
	if err != nil {
		log.Err(err)
	}
}

var (
	// This is here because oapi-codegen will not generate params using $ref
	apiKeySchema    = openapi3.NewStringSchema().WithPattern("^[0-9A-Z]{20}$")
	keyGetParam     = openapi3.NewQueryParameter("key").WithSchema(apiKeySchema).WithRequired(true)
	keyPostParam    = openapi3.NewHeaderParameter("key").WithSchema(apiKeySchema).WithRequired(true)
	offsetParam     = openapi3.NewQueryParameter("offset").WithSchema(openapi3.NewIntegerSchema().WithDefault(0).WithMin(0))
	limitParam      = openapi3.NewQueryParameter("limit").WithSchema(openapi3.NewIntegerSchema().WithDefault(10).WithMin(1).WithMax(100))
	orderSortParam  = openapi3.NewQueryParameter("sort").WithSchema(openapi3.NewStringSchema())
	orderOrderParam = openapi3.NewQueryParameter("order").WithSchema(openapi3.NewStringSchema().WithEnum([]string{"asc", "desc"}))

	// Schemas
	priceSchema = &openapi3.Schema{
		Required: []string{"currency", "initial", "final", "discountPercent", "individual"},
		Properties: map[string]*openapi3.SchemaRef{
			"currency":        {Value: openapi3.NewStringSchema()},
			"initial":         {Value: openapi3.NewInt32Schema()},
			"final":           {Value: openapi3.NewInt32Schema()},
			"discountPercent": {Value: openapi3.NewInt32Schema()},
			"individual":      {Value: openapi3.NewInt32Schema()},
		},
	}
)

var Swagger = &openapi3.Swagger{
	OpenAPI: "3.0.0",
	Servers: []*openapi3.Server{
		{URL: config.Config.GameDBDomain.Get() + "/api"},
	},
	Info: &openapi3.Info{
		Title:   "Steam DB API",
		Version: "1",
		Contact: &openapi3.Contact{
			Name:  "Jleagle",
			URL:   "https://gamedb.online/contact",
			Email: "jimeagle@gmail.com",
		},
	},
	Security: openapi3.SecurityRequirements{
		{
			"key-header": []string{},
			"key-query":  []string{},
			"key-cookie": []string{},
		},
	},
	Components: openapi3.Components{
		SecuritySchemes: map[string]*openapi3.SecuritySchemeRef{
			"key-header": {Value: openapi3.NewSecurityScheme().WithName("key").WithType("apiKey").WithIn("header")},
			"key-query":  {Value: openapi3.NewSecurityScheme().WithName("key").WithType("apiKey").WithIn("query")},
		},
		Schemas: map[string]*openapi3.SchemaRef{
			"pagination-schema": {
				Value: &openapi3.Schema{
					Required: []string{"offset", "limit", "total", "pagesTotal", "pagesCurrent"},
					Properties: map[string]*openapi3.SchemaRef{
						"offset":       {Value: openapi3.NewInt64Schema()},
						"limit":        {Value: openapi3.NewInt64Schema()},
						"total":        {Value: openapi3.NewInt64Schema()},
						"pagesTotal":   {Value: openapi3.NewIntegerSchema()},
						"pagesCurrent": {Value: openapi3.NewIntegerSchema()},
					},
				},
			},
			"app-schema": {
				Value: &openapi3.Schema{
					Required: []string{"id", "name", "tags", "genres", "categories", "developers", "publishers", "prices", "players_max", "players_week_max", "players_week_avg", "release_date", "reviews_positive", "reviews_negative", "reviews_score", "metacritic_score"},
					Properties: map[string]*openapi3.SchemaRef{
						"id":               {Value: openapi3.NewIntegerSchema()},
						"name":             {Value: openapi3.NewStringSchema()},
						"tags":             {Value: openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema())},
						"genres":           {Value: openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema())},
						"categories":       {Value: openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema())},
						"developers":       {Value: openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema())},
						"publishers":       {Value: openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema())},
						"prices":           {Value: openapi3.NewArraySchema().WithItems(priceSchema)},
						"players_max":      {Value: openapi3.NewIntegerSchema()},
						"players_week_max": {Value: openapi3.NewIntegerSchema()},
						"players_week_avg": {Value: openapi3.NewFloat64Schema().WithFormat("double")},
						"release_date":     {Value: openapi3.NewInt64Schema()},
						"reviews_positive": {Value: openapi3.NewIntegerSchema()},
						"reviews_negative": {Value: openapi3.NewIntegerSchema()},
						"reviews_score":    {Value: openapi3.NewFloat64Schema().WithFormat("double")},
						"metacritic_score": {Value: openapi3.NewInt32Schema()},
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
				Value: priceSchema,
			},
		},
		Responses: map[string]*openapi3.ResponseRef{
			"message-response": {
				Value: &openapi3.Response{
					ExtensionProps: openapi3.ExtensionProps{},
					Description:    "Message",
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/message-schema",
					}),
				},
			},
			"pagination-response": {
				Value: &openapi3.Response{
					Description: "Page information",
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/pagination-schema",
					}),
				},
			},
			"app-response": {
				Value: &openapi3.Response{
					Description: "An app",
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/app-schema",
					}),
				},
			},
			"apps-response": {
				Value: &openapi3.Response{
					Description: "List of apps",
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
			"player-response": {
				Value: &openapi3.Response{
					Description: "A player",
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/player-schema",
					}),
				},
			},
			"players-response": {
				Value: &openapi3.Response{
					Description: "List of players",
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
		"/apps": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Summary: "List Apps",
				Parameters: openapi3.Parameters{
					{Value: keyGetParam},
					{Value: offsetParam},
					{Value: limitParam},
					{Value: orderSortParam},
					{Value: orderOrderParam},
					{Value: openapi3.NewQueryParameter("ids").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema()).WithMaxItems(100))},
					{Value: openapi3.NewQueryParameter("tags").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema()).WithMaxItems(10))},
					{Value: openapi3.NewQueryParameter("genres").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema()).WithMaxItems(10))},
					{Value: openapi3.NewQueryParameter("categories").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema()).WithMaxItems(10))},
					{Value: openapi3.NewQueryParameter("developers").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema()).WithMaxItems(10))},
					{Value: openapi3.NewQueryParameter("publishers").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewIntegerSchema()).WithMaxItems(10))},
					{Value: openapi3.NewQueryParameter("platforms").WithSchema(openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema()).WithMaxItems(3))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/apps-response",
					},
				},
			},
		},
		"/apps/{id}": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Summary: "Retrieve App",
				Parameters: openapi3.Parameters{
					{Value: keyGetParam},
					{Value: openapi3.NewPathParameter("id").WithRequired(true).WithSchema(openapi3.NewInt32Schema().WithMin(1))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/app-response",
					},
				},
			},
		},
		"/players": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Summary: "List Players",
				Parameters: openapi3.Parameters{
					{Value: keyGetParam},
					{Value: offsetParam},
					{Value: limitParam},
					{Value: orderSortParam},
					{Value: orderOrderParam},
					{Value: openapi3.NewQueryParameter("continent").WithSchema(openapi3.NewArraySchema().WithMaxItems(3).WithItems(openapi3.NewStringSchema().WithMaxLength(2)))},
					{Value: openapi3.NewQueryParameter("country").WithSchema(openapi3.NewArraySchema().WithMaxItems(3).WithItems(openapi3.NewStringSchema().WithMaxLength(2)))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/players-response",
					},
				},
			},
		},
		"/players/{id}": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Summary: "Retrieve Player",
				Parameters: openapi3.Parameters{
					{Value: keyGetParam},
					{Value: openapi3.NewPathParameter("id").WithRequired(true).WithSchema(openapi3.NewInt64Schema().WithMin(1))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/player-response",
					},
				},
			},
			Post: &openapi3.Operation{
				Summary: "Update Player",
				Parameters: openapi3.Parameters{
					{Value: keyPostParam},
					{Value: openapi3.NewPathParameter("id").WithRequired(true).WithSchema(openapi3.NewInt64Schema().WithMaxLength(2))},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Ref: "#/components/responses/message-response",
					},
				},
			},
		},
		// "/app - players",
		// "/app - price changes",
		// "/articles",
		// "/bundles",
		// "/bundles",
		// "/bundles/{id}",
		// "/changes",
		// "/groups"
		// "/packages"
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
