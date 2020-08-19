package api

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/Jleagle/steam-go/steamapi"
	githubHelper "github.com/gamedb/gamedb/pkg/github"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/go-github/v28/github"
	"go.uber.org/zap"
)

var steamJSONLock sync.Mutex

func GetSteamJSON() *openapi3.Swagger {

	var groups = map[string][]string{}

	var swagger = &openapi3.Swagger{
		Servers: []*openapi3.Server{
			{URL: "https://api.steampowered.com/"},
		},
		Info: &openapi3.Info{
			Title: "Steam API",
		},
		ExternalDocs: &openapi3.ExternalDocs{
			URL: "https://steamcommunity.com/dev",
		},
		Paths: openapi3.Paths{},
		ExtensionProps: openapi3.ExtensionProps{
			Extensions: map[string]interface{}{
				"x-tagGroups": []interface{}{},
			},
		},
	}

	steamResp, err := steam.GetSteam().GetSupportedAPIList()
	err = steam.AllowSteamCodes(err)
	if err != nil {
		zap.S().Error(err)
		return swagger
	}

	for _, interfacex := range steamResp.Interfaces {
		addInterfaceToSwagger(swagger, &interfacex, groups)
	}

	// Put into a map to remove dupes from Github
	client, ctx := githubHelper.GetGithub()
	_, dirs, _, err := client.Repositories.GetContents(ctx, "SteamDatabase", "SteamTracking", "API", nil)
	if err != nil {
		zap.S().Error(err)
		return swagger
	}

	var wg sync.WaitGroup
	for _, dir := range dirs {

		wg.Add(1)
		go func(dir *github.RepositoryContent) {

			defer wg.Done()

			if !strings.HasSuffix(*dir.DownloadURL, ".json") {
				return
			}

			body, _, err := helpers.GetWithTimeout(*dir.DownloadURL, 0)
			if err != nil {
				zap.S().Error(err)
				return
			}

			i := steamapi.APIInterface{}
			err = helpers.Unmarshal(body, &i)
			if err != nil {
				zap.S().Error(err)
				return
			}

			addInterfaceToSwagger(swagger, &i, groups)
		}(dir)
	}
	wg.Wait()

	// Set groups
	var is []interface{}
	for k, v := range groups {

		is = append(is, map[string]interface{}{
			"name": k,
			"tags": v,
		})
	}
	// swagger.ExtensionProps.Extensions["x-tagGroups"] = is

	return swagger
}

var groupsMap = map[string]*regexp.Regexp{
	"Dota 2":          regexp.MustCompile(`^IDOTA2`),
	"Economy":         regexp.MustCompile(`^IEcon`),
	"Steam":           regexp.MustCompile(`^ISteam`),
	"Team Fortress 2": regexp.MustCompile(`^ITF`),
}

func addInterfaceToSwagger(swagger *openapi3.Swagger, interfacex *steamapi.APIInterface, groups map[string][]string) {

	steamJSONLock.Lock()
	defer steamJSONLock.Unlock()

	var found bool
	for k, v := range groupsMap {
		if v.MatchString(interfacex.Name) {
			found = true
			groups[k] = append(groups[k], interfacex.Name)
			break
		}
	}
	if !found {
		groups["Other"] = append(groups["Other"], interfacex.Name)
	}

	for _, method := range interfacex.Methods {

		operation := &openapi3.Operation{
			Tags:    []string{interfacex.Name},
			Summary: method.Name,
		}

		for _, param := range method.Parameters {

			paramx := openapi3.
				NewQueryParameter(param.Name).
				WithDescription(param.Description).
				WithRequired(param.Optional)

			switch param.Type {
			case "int32", "uint32":
				paramx.WithSchema(openapi3.NewInt32Schema())
			case "uint64":
				paramx.WithSchema(openapi3.NewInt64Schema())
			case "string", "rawbinary", "{message}", "{enum}":
				paramx.WithSchema(openapi3.NewStringSchema())
			case "bool":
				paramx.WithSchema(openapi3.NewBoolSchema())
			case "float":
				paramx.WithSchema(openapi3.NewFloat64Schema())
			default:
				zap.S().Warn("new param type", param.Type, interfacex.Name)
			}

			operation.Parameters = append(operation.Parameters, &openapi3.ParameterRef{Value: paramx})
		}

		path := "/" + interfacex.Name + "/" + method.Name + "/v" + strconv.Itoa(method.Version)

		switch method.HTTPmethod {
		case http.MethodGet:
			swagger.Paths[path] = &openapi3.PathItem{Get: operation}
		case http.MethodPost:
			swagger.Paths[path] = &openapi3.PathItem{Get: operation}
		default:
			zap.L().Warn("new http method")
		}
	}
}
