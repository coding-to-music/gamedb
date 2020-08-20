// Package generated provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package generated

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi"
	"net/http"
	"strings"
)

// AppSchema defines model for app-schema.
type AppSchema struct {
	Categories      []int   `json:"categories"`
	Developers      []int   `json:"developers"`
	Genres          []int   `json:"genres"`
	Id              int     `json:"id"`
	MetacriticScore int32   `json:"metacritic_score"`
	Name            string  `json:"name"`
	PlayersMax      int     `json:"players_max"`
	PlayersWeekAvg  float64 `json:"players_week_avg"`
	PlayersWeekMax  int     `json:"players_week_max"`
	Prices          []struct {
		Currency        string `json:"currency"`
		DiscountPercent int32  `json:"discountPercent"`
		Final           int32  `json:"final"`
		Free            bool   `json:"free"`
		Individual      int32  `json:"individual"`
		Initial         int32  `json:"initial"`
	} `json:"prices"`
	Publishers      []int   `json:"publishers"`
	ReleaseDate     int64   `json:"release_date"`
	ReviewsNegative int     `json:"reviews_negative"`
	ReviewsPositive int     `json:"reviews_positive"`
	ReviewsScore    float64 `json:"reviews_score"`
	Tags            []int   `json:"tags"`
}

// MessageSchema defines model for message-schema.
type MessageSchema struct {
	Message string `json:"message"`
}

// PaginationSchema defines model for pagination-schema.
type PaginationSchema struct {
	Limit        int64 `json:"limit"`
	Offset       int64 `json:"offset"`
	PagesCurrent int64 `json:"pagesCurrent"`
	PagesTotal   int64 `json:"pagesTotal"`
	Total        int64 `json:"total"`
}

// PlayerSchema defines model for player-schema.
type PlayerSchema struct {
	Avatar    string `json:"avatar"`
	Badges    int    `json:"badges"`
	Comments  int    `json:"comments"`
	Continent string `json:"continent"`
	Country   string `json:"country"`
	Friends   int    `json:"friends"`
	Games     int    `json:"games"`
	Groups    int    `json:"groups"`
	Id        string `json:"id"`
	Level     int    `json:"level"`
	Name      string `json:"name"`
	Playtime  int    `json:"playtime"`
	State     string `json:"state"`
	VanityUrl string `json:"vanity_url"`
}

// LimitParam defines model for limit-param.
type LimitParam int

// OffsetParam defines model for offset-param.
type OffsetParam int

// OrderParamDesc defines model for order-param-desc.
type OrderParamDesc string

// List of OrderParamDesc
const (
	OrderParamDesc_asc  OrderParamDesc = "asc"
	OrderParamDesc_desc OrderParamDesc = "desc"
)

// AppResponse defines model for app-response.
type AppResponse AppSchema

// AppsResponse defines model for apps-response.
type AppsResponse struct {
	Apps       []AppSchema      `json:"apps"`
	Pagination PaginationSchema `json:"pagination"`
}

// MessageResponse defines model for message-response.
type MessageResponse MessageSchema

// PlayerResponse defines model for player-response.
type PlayerResponse PlayerSchema

// PlayersResponse defines model for players-response.
type PlayersResponse struct {
	Pagination PaginationSchema `json:"pagination"`
	Players    []PlayerSchema   `json:"players"`
}

// GetGamesParams defines parameters for GetGames.
type GetGamesParams struct {
	Key        string          `json:"key"`
	Offset     *OffsetParam    `json:"offset,omitempty"`
	Limit      *LimitParam     `json:"limit,omitempty"`
	Order      *OrderParamDesc `json:"order,omitempty"`
	Sort       *string         `json:"sort,omitempty"`
	Ids        *[]int          `json:"ids,omitempty"`
	Tags       *[]int          `json:"tags,omitempty"`
	Genres     *[]int          `json:"genres,omitempty"`
	Categories *[]int          `json:"categories,omitempty"`
	Developers *[]int          `json:"developers,omitempty"`
	Publishers *[]int          `json:"publishers,omitempty"`
	Platforms  *[]string       `json:"platforms,omitempty"`
}

// GetGamesIdParams defines parameters for GetGamesId.
type GetGamesIdParams struct {
	Key string `json:"key"`
}

// GetPlayersParams defines parameters for GetPlayers.
type GetPlayersParams struct {
	Key       string          `json:"key"`
	Offset    *OffsetParam    `json:"offset,omitempty"`
	Limit     *LimitParam     `json:"limit,omitempty"`
	Order     *OrderParamDesc `json:"order,omitempty"`
	Sort      *string         `json:"sort,omitempty"`
	Continent *[]string       `json:"continent,omitempty"`
	Country   *[]string       `json:"country,omitempty"`
}

// GetPlayersIdParams defines parameters for GetPlayersId.
type GetPlayersIdParams struct {
	Key string `json:"key"`
}

// PostPlayersIdParams defines parameters for PostPlayersId.
type PostPlayersIdParams struct {
	Key string `json:"key"`
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// List Apps
	// (GET /games)
	GetGames(w http.ResponseWriter, r *http.Request, params GetGamesParams)
	// Retrieve App
	// (GET /games/{id})
	GetGamesId(w http.ResponseWriter, r *http.Request, id int32, params GetGamesIdParams)
	// List Players
	// (GET /players)
	GetPlayers(w http.ResponseWriter, r *http.Request, params GetPlayersParams)
	// Retrieve Player
	// (GET /players/{id})
	GetPlayersId(w http.ResponseWriter, r *http.Request, id int64, params GetPlayersIdParams)
	// Update Player
	// (POST /players/{id})
	PostPlayersId(w http.ResponseWriter, r *http.Request, id int64, params PostPlayersIdParams)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetGames operation middleware
func (siw *ServerInterfaceWrapper) GetGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	ctx = context.WithValue(ctx, "key-header.Scopes", []string{""})

	ctx = context.WithValue(ctx, "key-query.Scopes", []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetGamesParams

	// ------------- Required query parameter "key" -------------
	if paramValue := r.URL.Query().Get("key"); paramValue != "" {

	} else {
		http.Error(w, "Query argument key is required, but not found", http.StatusBadRequest)
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "key", r.URL.Query(), &params.Key)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter key: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "offset" -------------
	if paramValue := r.URL.Query().Get("offset"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "offset", r.URL.Query(), &params.Offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter offset: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "limit" -------------
	if paramValue := r.URL.Query().Get("limit"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "limit", r.URL.Query(), &params.Limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter limit: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "order" -------------
	if paramValue := r.URL.Query().Get("order"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "order", r.URL.Query(), &params.Order)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter order: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "sort" -------------
	if paramValue := r.URL.Query().Get("sort"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "sort", r.URL.Query(), &params.Sort)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter sort: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "ids" -------------
	if paramValue := r.URL.Query().Get("ids"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "ids", r.URL.Query(), &params.Ids)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter ids: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "tags" -------------
	if paramValue := r.URL.Query().Get("tags"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "tags", r.URL.Query(), &params.Tags)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter tags: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "genres" -------------
	if paramValue := r.URL.Query().Get("genres"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "genres", r.URL.Query(), &params.Genres)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter genres: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "categories" -------------
	if paramValue := r.URL.Query().Get("categories"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "categories", r.URL.Query(), &params.Categories)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter categories: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "developers" -------------
	if paramValue := r.URL.Query().Get("developers"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "developers", r.URL.Query(), &params.Developers)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter developers: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "publishers" -------------
	if paramValue := r.URL.Query().Get("publishers"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "publishers", r.URL.Query(), &params.Publishers)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter publishers: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "platforms" -------------
	if paramValue := r.URL.Query().Get("platforms"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "platforms", r.URL.Query(), &params.Platforms)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter platforms: %s", err), http.StatusBadRequest)
		return
	}

	siw.Handler.GetGames(w, r.WithContext(ctx), params)
}

// GetGamesId operation middleware
func (siw *ServerInterfaceWrapper) GetGamesId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "id" -------------
	var id int32

	err = runtime.BindStyledParameter("simple", false, "id", chi.URLParam(r, "id"), &id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter id: %s", err), http.StatusBadRequest)
		return
	}

	ctx = context.WithValue(ctx, "key-header.Scopes", []string{""})

	ctx = context.WithValue(ctx, "key-query.Scopes", []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetGamesIdParams

	// ------------- Required query parameter "key" -------------
	if paramValue := r.URL.Query().Get("key"); paramValue != "" {

	} else {
		http.Error(w, "Query argument key is required, but not found", http.StatusBadRequest)
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "key", r.URL.Query(), &params.Key)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter key: %s", err), http.StatusBadRequest)
		return
	}

	siw.Handler.GetGamesId(w, r.WithContext(ctx), id, params)
}

// GetPlayers operation middleware
func (siw *ServerInterfaceWrapper) GetPlayers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	ctx = context.WithValue(ctx, "key-header.Scopes", []string{""})

	ctx = context.WithValue(ctx, "key-query.Scopes", []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetPlayersParams

	// ------------- Required query parameter "key" -------------
	if paramValue := r.URL.Query().Get("key"); paramValue != "" {

	} else {
		http.Error(w, "Query argument key is required, but not found", http.StatusBadRequest)
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "key", r.URL.Query(), &params.Key)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter key: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "offset" -------------
	if paramValue := r.URL.Query().Get("offset"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "offset", r.URL.Query(), &params.Offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter offset: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "limit" -------------
	if paramValue := r.URL.Query().Get("limit"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "limit", r.URL.Query(), &params.Limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter limit: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "order" -------------
	if paramValue := r.URL.Query().Get("order"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "order", r.URL.Query(), &params.Order)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter order: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "sort" -------------
	if paramValue := r.URL.Query().Get("sort"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "sort", r.URL.Query(), &params.Sort)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter sort: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "continent" -------------
	if paramValue := r.URL.Query().Get("continent"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "continent", r.URL.Query(), &params.Continent)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter continent: %s", err), http.StatusBadRequest)
		return
	}

	// ------------- Optional query parameter "country" -------------
	if paramValue := r.URL.Query().Get("country"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "country", r.URL.Query(), &params.Country)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter country: %s", err), http.StatusBadRequest)
		return
	}

	siw.Handler.GetPlayers(w, r.WithContext(ctx), params)
}

// GetPlayersId operation middleware
func (siw *ServerInterfaceWrapper) GetPlayersId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "id" -------------
	var id int64

	err = runtime.BindStyledParameter("simple", false, "id", chi.URLParam(r, "id"), &id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter id: %s", err), http.StatusBadRequest)
		return
	}

	ctx = context.WithValue(ctx, "key-header.Scopes", []string{""})

	ctx = context.WithValue(ctx, "key-query.Scopes", []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetPlayersIdParams

	// ------------- Required query parameter "key" -------------
	if paramValue := r.URL.Query().Get("key"); paramValue != "" {

	} else {
		http.Error(w, "Query argument key is required, but not found", http.StatusBadRequest)
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "key", r.URL.Query(), &params.Key)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter key: %s", err), http.StatusBadRequest)
		return
	}

	siw.Handler.GetPlayersId(w, r.WithContext(ctx), id, params)
}

// PostPlayersId operation middleware
func (siw *ServerInterfaceWrapper) PostPlayersId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "id" -------------
	var id int64

	err = runtime.BindStyledParameter("simple", false, "id", chi.URLParam(r, "id"), &id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid format for parameter id: %s", err), http.StatusBadRequest)
		return
	}

	ctx = context.WithValue(ctx, "key-header.Scopes", []string{""})

	ctx = context.WithValue(ctx, "key-query.Scopes", []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params PostPlayersIdParams

	headers := r.Header

	// ------------- Required header parameter "key" -------------
	if valueList, found := headers[http.CanonicalHeaderKey("key")]; found {
		var Key string
		n := len(valueList)
		if n != 1 {
			http.Error(w, fmt.Sprintf("Expected one value for key, got %d", n), http.StatusBadRequest)
			return
		}

		err = runtime.BindStyledParameter("simple", false, "key", valueList[0], &Key)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter key: %s", err), http.StatusBadRequest)
			return
		}

		params.Key = Key

	} else {
		http.Error(w, fmt.Sprintf("Header parameter key is required, but not found: %s", err), http.StatusBadRequest)
		return
	}

	siw.Handler.PostPlayersId(w, r.WithContext(ctx), id, params)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerFromMux(si, chi.NewRouter())
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	r.Group(func(r chi.Router) {
		r.Get("/games", wrapper.GetGames)
	})
	r.Group(func(r chi.Router) {
		r.Get("/games/{id}", wrapper.GetGamesId)
	})
	r.Group(func(r chi.Router) {
		r.Get("/players", wrapper.GetPlayers)
	})
	r.Group(func(r chi.Router) {
		r.Get("/players/{id}", wrapper.GetPlayersId)
	})
	r.Group(func(r chi.Router) {
		r.Post("/players/{id}", wrapper.PostPlayersId)
	})

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file.
func GetSwagger() (*openapi3.Swagger, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error loading Swagger: %s", err)
	}
	return swagger, nil
}
