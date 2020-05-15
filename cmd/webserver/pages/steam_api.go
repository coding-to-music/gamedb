package pages

import (
	"encoding/gob"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	steamHelper "github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi"
	"github.com/google/go-github/v28/github"
)

func init() {
	gob.Register(&Interfaces{})
}

func SteamAPIRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", steamAPIHandler)
	r.Get("/openapi.json", steamAPISwaggerHandler)
	return r
}

func steamAPISwaggerHandler(w http.ResponseWriter, r *http.Request) {

	swagger := openapi3.Swagger{}

	b, err := swagger.MarshalJSON()
	log.Err(err)

	_, err = w.Write(b)
	log.Err(err)
}

func steamAPIHandler(w http.ResponseWriter, r *http.Request) {

	t := steamAPITemplate{}
	t.fill(w, r, "Steam API", "Steam API documentation")

	var err error
	var interfaces = Interfaces{}

	retrieve := func() (interface{}, error) {
		err = interfaces.addDocumented(w, r)
		if err != nil {
			return nil, err
		}

		err = interfaces.addUndocumented()
		if err != nil {
			return nil, err
		}

		return interfaces, nil
	}

	err = helpers.GetSetCache("steam-api", time.Hour*24, retrieve, &interfaces)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "An error occurred"})
		return
	}

	t.Interfaces = interfaces
	t.Description = template.HTML(strconv.Itoa(t.Count())) + " of the known Steam web API endpoints."

	returnTemplate(w, r, "steam_api", t)
}

type steamAPITemplate struct {
	GlobalTemplate
	Interfaces Interfaces
}

func (t steamAPITemplate) Count() (count int) {
	for _, v := range t.Interfaces {
		for _, vv := range v {
			for _, vvv := range vv.Versions {
				for range vvv {
					count++
				}
			}
		}
	}
	return count
}

type Interfaces map[string]Methods

type Methods map[string]Method

type Method struct {
	Versions   map[int]map[string]map[string]Param
	Documented bool
	Publisher  bool
}

type Param struct {
	Type        string `json:"type"`
	Optional    bool   `json:"optional"`
	Description string `json:"description"`
}

var addMutex sync.Mutex

func (interfaces *Interfaces) addInterface(in steamapi.APIInterface, documented bool) {

	addMutex.Lock()
	defer addMutex.Unlock()

	for _, method := range in.Methods {

		for _, param := range method.Parameters {

			if (*interfaces)[in.Name] == nil {
				(*interfaces)[in.Name] = map[string]Method{}
			}

			if (*interfaces)[in.Name][method.Name].Versions == nil {
				(*interfaces)[in.Name][method.Name] = Method{
					Documented: documented,
					Versions:   map[int]map[string]map[string]Param{},
				}
			}

			if (*interfaces)[in.Name][method.Name].Versions[method.Version] == nil {
				(*interfaces)[in.Name][method.Name].Versions[method.Version] = make(map[string]map[string]Param)
			}

			if (*interfaces)[in.Name][method.Name].Versions[method.Version][method.HTTPmethod] == nil {
				(*interfaces)[in.Name][method.Name].Versions[method.Version][method.HTTPmethod] = make(map[string]Param)
			}

			(*interfaces)[in.Name][method.Name].Versions[method.Version][method.HTTPmethod][param.Name] = Param{
				Type:        param.Type,
				Optional:    param.Optional,
				Description: param.Description,
			}
		}
	}
}

func (interfaces *Interfaces) addDocumented(w http.ResponseWriter, r *http.Request) (err error) {

	steamResp, b, err := steamHelper.GetSteam().GetSupportedAPIList()
	err = steamHelper.AllowSteamCodes(err, b, nil)
	if err != nil {
		return err
	}

	// Put into a map to remove dupes from Github
	for _, v := range steamResp.Interfaces {
		interfaces.addInterface(v, true)
	}

	return nil
}

func (interfaces *Interfaces) addUndocumented() (err error) {

	client, ctx := helpers.GetGithub()
	_, dirs, _, err := client.Repositories.GetContents(ctx, "SteamDatabase", "SteamTracking", "API", nil)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, dir := range dirs {

		wg.Add(1)
		go func(dir *github.RepositoryContent) {

			defer wg.Done()

			if !strings.HasSuffix(*dir.DownloadURL, ".json") {
				return
			}

			resp, err := helpers.GetWithTimeout(*dir.DownloadURL, 0)
			if err != nil {
				log.Err(err)
				return
			}

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Err(err)
				return
			}

			i := steamapi.APIInterface{}
			err = helpers.Unmarshal(b, &i)
			if err != nil {
				log.Err(err)
				return
			}

			interfaces.addInterface(i, false)
		}(dir)
	}

	wg.Wait()

	return nil
}
