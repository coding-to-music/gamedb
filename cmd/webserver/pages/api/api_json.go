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
	float0   float64 = 0
	float1   float64 = 1
	float100 float64 = 100

	// This is here because oapi-codegen wont generate params using $ref
	test = []*openapi3.ParameterRef{
		{
			Value: &openapi3.Parameter{
				In:          openapi3.ParameterInQuery,
				Name:        "offset",
				Required:    false,
				Description: "Offset",
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:    "integer",
						Default: 0,
						Min:     &float0,
					},
				},
			},
		},
		{
			Value: &openapi3.Parameter{
				In:          openapi3.ParameterInQuery,
				Name:        "limit",
				Required:    false,
				Description: "Limit",
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:    "integer",
						Default: 10,
						Min:     &float1,
						Max:     &float100,
					},
				},
			},
		},
	}
)

var Swagger = &openapi3.Swagger{
	OpenAPI: "3.0.0",
	Servers: []*openapi3.Server{
		{URL: config.Config.GameDBDomain.Get() + "/api"},
	},
	Info: openapi3.Info{
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
			"key-header": {
				Value: &openapi3.SecurityScheme{
					Name: "key",
					Type: "apiKey",
					In:   "header",
				},
			},
			"key-query": {
				Value: &openapi3.SecurityScheme{
					Name: "key",
					Type: "apiKey",
					In:   "query",
				},
			},
			"key-cookie": {
				Value: &openapi3.SecurityScheme{
					Name: "key",
					Type: "apiKey",
					In:   "cookie",
				},
			},
		},
		Schemas: map[string]*openapi3.SchemaRef{
			"pagination-schema": {
				Value: &openapi3.Schema{
					Required: []string{"offset", "limit", "rowsTotal", "rowsFiltered", "pagesTotal", "pagescurrent"},
					Properties: map[string]*openapi3.SchemaRef{
						"offset":       {Value: openapi3.NewIntegerSchema()},
						"limit":        {Value: openapi3.NewIntegerSchema()},
						"rowsTotal":    {Value: openapi3.NewIntegerSchema()},
						"rowsFiltered": {Value: openapi3.NewIntegerSchema()},
						"pagesTotal":   {Value: openapi3.NewIntegerSchema()},
						"pagescurrent": {Value: openapi3.NewIntegerSchema()},
					},
				},
			},
			"app-schema": {
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
					Properties: map[string]*openapi3.SchemaRef{
						"id":   {Value: openapi3.NewIntegerSchema()},
						"name": {Value: openapi3.NewStringSchema()},
					},
				},
			},
			"error-schema": {
				Value: &openapi3.Schema{
					Required: []string{"code", "message"},
					Properties: map[string]*openapi3.SchemaRef{
						"code":    {Value: openapi3.NewIntegerSchema()},
						"message": {Value: openapi3.NewStringSchema()},
					},
				},
			},
			"succcess-schema": {
				Value: &openapi3.Schema{
					Required: []string{"message"},
					Properties: map[string]*openapi3.SchemaRef{
						"message": {Value: openapi3.NewStringSchema()},
					},
				},
			},
		},
		Responses: map[string]*openapi3.ResponseRef{
			"pagination-response": {
				Value: &openapi3.Response{
					Description: "Page information",
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/pagination-schema",
					}),
				},
			},
			"error": {
				Value: &openapi3.Response{
					Description: "An error",
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/error-schema",
					}),
				},
			},
			"app": {
				Value: &openapi3.Response{
					Description: "An app",
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/app-schema",
					}),
				},
			},
			"apps": {
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
			"player": {
				Value: &openapi3.Response{
					Description: "A player",
					Content: openapi3.NewContentWithJSONSchemaRef(&openapi3.SchemaRef{
						Ref: "#/components/schemas/player-schema",
					}),
				},
			},
			"players": {
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
		Tags: openapi3.Tags{
			{Name: "apps", Description: ""},
			{Name: "players", Description: ""},
		},
	},
	Paths: openapi3.Paths{
		"/apps": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Summary: "List apps",
				Tags:    []string{"apps"},
				Parameters: append(openapi3.Parameters{
					{
						Value: &openapi3.Parameter{
							In:     openapi3.ParameterInQuery,
							Name:   "ids",
							Schema: openapi3.NewArraySchema().WithFormat("integer").WithMaxItems(100).NewRef(),
						},
					},
					{
						Value: &openapi3.Parameter{
							In:     openapi3.ParameterInQuery,
							Name:   "tags",
							Schema: openapi3.NewArraySchema().WithFormat("integer").WithMaxItems(10).NewRef(),
						},
					},
				}, test...),
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Value: &openapi3.Response{
							ExtensionProps: openapi3.ExtensionProps{},
							Description:    "List of apps",
							Content: openapi3.Content{
								"application/json": &openapi3.MediaType{
									Schema: &openapi3.SchemaRef{
										Ref: "#/components/schemas/app-schema",
									},
								},
							},
						},
					},
				},
			},
		},
		"/apps/{id}": &openapi3.PathItem{
			Get: &openapi3.Operation{
				Summary: "Retrieve app",
				Tags:    []string{"apps"},
				Parameters: openapi3.Parameters{
					{
						Value: &openapi3.Parameter{
							Required: true,
							In:       openapi3.ParameterInPath,
							Name:     "id",
							Schema:   openapi3.NewInt32Schema().WithMin(1).NewRef(),
						},
					},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Value: &openapi3.Response{
							ExtensionProps: openapi3.ExtensionProps{},
							Description:    "An app",
							Content: openapi3.Content{
								"application/json": &openapi3.MediaType{
									Schema: &openapi3.SchemaRef{
										Ref: "#/components/schemas/app-schema",
									},
								},
							},
						},
					},
				},
			},
		},
		"/players/{id}": &openapi3.PathItem{
			Post: &openapi3.Operation{
				Summary: "Update a player",
				Tags:    []string{"players"},
				Parameters: openapi3.Parameters{
					{
						Value: &openapi3.Parameter{
							Required: true,
							In:       openapi3.ParameterInPath,
							Name:     "id",
							Schema:   openapi3.NewInt64Schema().WithMaxLength(2).NewRef(),
						},
					},
				},
				Responses: map[string]*openapi3.ResponseRef{
					"200": {
						Value: &openapi3.Response{
							ExtensionProps: openapi3.ExtensionProps{},
							Description:    "Success",
							Content: openapi3.Content{
								"application/json": &openapi3.MediaType{
									Schema: &openapi3.SchemaRef{
										Ref: "#/components/schemas/succcess-schema",
									},
								},
							},
						},
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
