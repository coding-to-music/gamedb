package web

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/google/go-github/github"
)

var addMutex sync.Mutex

func steamAPIHandler(w http.ResponseWriter, r *http.Request) {

	t := steamAPITemplate{}
	t.fill(w, r, "Steam API", "")
	t.Interfaces = Interfaces{}
	t.addDocumented(w, r)
	t.addUndocumented()
	t.Description = template.HTML(strconv.Itoa(t.Count())) + " of the known Steam web API endpoints."

	err := returnTemplate(w, r, "steam_api", t)
	log.Err(err, r)
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

func (t *steamAPITemplate) addDocumented(w http.ResponseWriter, r *http.Request) {

	steamResp, b, err := helpers.GetSteam().GetSupportedAPIList()
	err = helpers.HandleSteamStoreErr(err, b, nil)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Can't talk to Steam"})
		return
	}

	// Put into a map to remove dupes from Github
	for _, v := range steamResp.Interfaces {
		t.Interfaces.addInterface(v, true)
	}
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

func (t *steamAPITemplate) addUndocumented() {

	client, ctx := helpers.GetGithub()
	_, dir, _, err := client.Repositories.GetContents(ctx, "SteamDatabase", "SteamTracking", "API", nil)
	log.Err(err)

	var wg sync.WaitGroup
	for _, v := range dir {

		wg.Add(1)
		go func(v *github.RepositoryContent) {

			defer wg.Done()

			if !strings.HasSuffix(*v.DownloadURL, ".json") {
				return
			}

			resp, err := http.Get(*v.DownloadURL)
			if err != nil {
				log.Err(err)
				return
			}

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Err(err)
				return
			}

			i := steam.APIInterface{}
			err = helpers.Unmarshal(b, &i)
			if err != nil {
				log.Err(err)
				return
			}

			t.Interfaces.addInterface(i, false)
		}(v)
	}

	wg.Wait()
}

func (i *Interfaces) addInterface(in steam.APIInterface, documented bool) {

	addMutex.Lock()
	defer addMutex.Unlock()

	for _, method := range in.Methods {

		for _, param := range method.Parameters {

			if (*i)[in.Name] == nil {
				(*i)[in.Name] = map[string]Method{}
			}

			if (*i)[in.Name][method.Name].Versions == nil {
				(*i)[in.Name][method.Name] = Method{
					Documented: documented,
					Versions:   map[int]map[string]map[string]Param{},
				}
			}

			if (*i)[in.Name][method.Name].Versions[method.Version] == nil {
				(*i)[in.Name][method.Name].Versions[method.Version] = make(map[string]map[string]Param)
			}

			if (*i)[in.Name][method.Name].Versions[method.Version][method.HTTPmethod] == nil {
				(*i)[in.Name][method.Name].Versions[method.Version][method.HTTPmethod] = make(map[string]Param)
			}

			(*i)[in.Name][method.Name].Versions[method.Version][method.HTTPmethod][param.Name] = Param{
				Type:        param.Type,
				Optional:    param.Optional,
				Description: param.Description,
			}
		}
	}
}
