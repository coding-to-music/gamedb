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
	Ids        *[]int32        `json:"ids,omitempty"`
	Tags       *[]int32        `json:"tags,omitempty"`
	Genres     *[]int32        `json:"genres,omitempty"`
	Categories *[]int32        `json:"categories,omitempty"`
	Developers *[]int32        `json:"developers,omitempty"`
	Publishers *[]int32        `json:"publishers,omitempty"`
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
	// List Games
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
