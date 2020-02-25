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
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// ErrorSchema defines model for error-schema.
type ErrorSchema struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// PaginationSchema defines model for pagination-schema.
type PaginationSchema struct {
	Limit        int64 `json:"limit"`
	Offset       int64 `json:"offset"`
	PagesCurrent *int  `json:"pagesCurrent,omitempty"`
	PagesTotal   int   `json:"pagesTotal"`
	RowsFiltered int64 `json:"rowsFiltered"`
	RowsTotal    int64 `json:"rowsTotal"`
}

// PlayerSchema defines model for player-schema.
type PlayerSchema struct {
	Id   *int    `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// SucccessSchema defines model for succcess-schema.
type SucccessSchema struct {
	Message string `json:"message"`
}

// App defines model for app.
type App AppSchema

// Apps defines model for apps.
type Apps struct {
	Apps       []AppSchema      `json:"apps"`
	Pagination PaginationSchema `json:"pagination"`
}

// Error defines model for error.
type Error ErrorSchema

// PaginationResponse defines model for pagination-response.
type PaginationResponse PaginationSchema

// Player defines model for player.
type Player PlayerSchema

// Players defines model for players.
type Players struct {
	Pagination PaginationSchema `json:"pagination"`
	Players    []PlayerSchema   `json:"players"`
}

// GetAppsParams defines parameters for GetApps.
type GetAppsParams struct {
	Ids  *[]interface{} `json:"ids,omitempty"`
	Tags *[]interface{} `json:"tags,omitempty"`

	// Offset
	Offset *int `json:"offset,omitempty"`

	// Limit
	Limit *int `json:"limit,omitempty"`
}

type ServerInterface interface {
	// List apps (GET /apps)
	GetApps(w http.ResponseWriter, r *http.Request)
	// Retrieve app (GET /apps/{id})
	GetAppsId(w http.ResponseWriter, r *http.Request)
	// Update a player (POST /players/{id})
	PostPlayersId(w http.ResponseWriter, r *http.Request)
}

// ParamsForGetApps operation parameters from context
func ParamsForGetApps(ctx context.Context) *GetAppsParams {
	return ctx.Value("GetAppsParams").(*GetAppsParams)
}

// GetApps operation middleware
func GetAppsCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var err error

		ctx = context.WithValue(ctx, "key-cookie.Scopes", []string{""})

		ctx = context.WithValue(ctx, "key-header.Scopes", []string{""})

		ctx = context.WithValue(ctx, "key-query.Scopes", []string{""})

		// Parameter object where we will unmarshal all parameters from the context
		var params GetAppsParams

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

		ctx = context.WithValue(ctx, "GetAppsParams", &params)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetAppsId operation middleware
func GetAppsIdCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var err error

		// ------------- Path parameter "id" -------------
		var id int32

		err = runtime.BindStyledParameter("simple", false, "id", chi.URLParam(r, "id"), &id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter id: %s", err), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, "id", id)

		ctx = context.WithValue(ctx, "key-header.Scopes", []string{""})

		ctx = context.WithValue(ctx, "key-query.Scopes", []string{""})

		ctx = context.WithValue(ctx, "key-cookie.Scopes", []string{""})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// PostPlayersId operation middleware
func PostPlayersIdCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var err error

		// ------------- Path parameter "id" -------------
		var id int64

		err = runtime.BindStyledParameter("simple", false, "id", chi.URLParam(r, "id"), &id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid format for parameter id: %s", err), http.StatusBadRequest)
			return
		}

		ctx = context.WithValue(ctx, "id", id)

		ctx = context.WithValue(ctx, "key-cookie.Scopes", []string{""})

		ctx = context.WithValue(ctx, "key-header.Scopes", []string{""})

		ctx = context.WithValue(ctx, "key-query.Scopes", []string{""})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerFromMux(si, chi.NewRouter())
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r *chi.Mux) http.Handler {
	r.Group(func(r chi.Router) {
		r.Use(GetAppsCtx)
		r.Get("/apps", si.GetApps)
	})
	r.Group(func(r chi.Router) {
		r.Use(GetAppsIdCtx)
		r.Get("/apps/{id}", si.GetAppsId)
	})
	r.Group(func(r chi.Router) {
		r.Use(PostPlayersIdCtx)
		r.Post("/players/{id}", si.PostPlayersId)
	})

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/8xXTW/jNhD9K8K0RyVyskWx0KlpixZpF2jQtCfDB1Yay9wVP5aksjEC//di+KEPW3Yc",
	"1Fj0Epvm8L3h45sh8wKVElpJlM5C+QIGrVbSoh8wremjUtKhdPGXllfMcSWLj1ZJ+s1WGxSMvn1rcA0l",
	"fFMMmEWYtQXT+ipG7na7HGq0leGakKCEO5kR2S4nBvsm0inQB25dptaEZvPsC3ebTLOGS78actBGaTSO",
	"9xv0n9yhsG/ZQQ5uqxFKYMawLY1HLK/gDJGDIDkY/NxxgzWUS5hk7HNczWg23iplgMYoc7Hz8minTywQ",
	"TvZ+lfxzsTzm1DpI5oE1mHG5VkYE2Siplm3xcnoEuFOCZJGx536bj6fO/E92miRwlrf3drdv71P+TFSn",
	"LJpiKCJSRkGujgnAa/ob0+DSYROklUzgaMY6w2VzkCGvIYauUmUcZapUjfNcAq1lzRl0HmKIX01r4hhx",
	"ywX3xgi+Ddzffwf5TCpqvbZ4brBmDdqfOmOi8Y5E/KUca+fnjfpif+GtQ7/Bs0hpSY/4avyegHF7edRk",
	"jLaXzCT3OKjiVle98y9mKzJsV1UVWnsU82ybDP4gWKw6w932kVAD0CfcXlVKfeIei1MFxWFKkEIGNZnm",
	"v6O/fWjlBlkdGp5fGYfnrPzcodn2C8Po9DrqEKyxUC5fUly8hvpxX/N0KtSbUzdklTclCsZbKOEjF8ia",
	"Fn9o6IfrSomB+7fWT0EOnaHYjXPalkXRMIH1P9dKtlxikUApKe5aWvjokIns5x+zu4d7yOEJjQ0d6cYX",
	"k0bJNIcS3l0vrhfeRm7jj6BIT4ImVBuds6/i+xpK+BXdHc3TAsMEOt9jl/PC8ZoCB9OMa8LbLgfBnu9D",
	"e75ZLA6b7jyuF/4NwHO40zb9R6q+Obq+NMdPrjXrWgflIgfBJRed8N8Pi/zweRbKe44olf4MD21CsOdA",
	"5KXqaW9maFf59B17u1h8pXfs9E3mm4cQjIorTLHgnlg7/dsuD74rXni9e8189/UR+5GHx+6DcfNxpsNj",
	"nnl3C/9bQdM/BhMp/0RnOD6hn5pVMzafXlCt7IyiD8q6hxB5eVX9tSfY8weUjdtAefuVdd2/tmbEfex8",
	"xJ66f+uaOcxYes8OAk+feukC81qNr67lanohpXG8ZpYr2rdF85SEHpp7WRStqli7UdaV7xfvbwrq07vV",
	"7t8AAAD//ylVryWoDgAA",
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
