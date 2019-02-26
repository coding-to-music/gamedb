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
	t.Fill(w, r, "Steam API", "")
	t.Interfaces = make(Interfaces)

	steamResp, _, err := helpers.GetSteam().GetSupportedAPIList()
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Can't talk to Steam"})
		return
	}

	// Put into a map to remove dupes from Github
	for _, v := range steamResp.Interfaces {
		t.Interfaces.addInterface(v)
	}

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

			t.Interfaces.addInterface(i)
		}(v)
	}

	wg.Wait()

	for _, v := range t.Interfaces {
		for _, vv := range v {
			for _, vvv := range vv {
				for range vvv {
					t.Count++
				}
			}
		}
	}

	t.Description = template.HTML(strconv.Itoa(t.Count)) + " of the known Steam web API endpoints."

	err = returnTemplate(w, r, "steam_api", t)
	log.Err(err, r)
}

type steamAPITemplate struct {
	GlobalTemplate
	Interfaces Interfaces
	Count      int
}

type Interfaces map[string]Interface

func (i *Interfaces) addInterface(in steam.APIInterface) {

	addMutex.Lock()
	defer addMutex.Unlock()

	for _, method := range in.Methods {

		for _, param := range method.Parameters {

			if (*i)[in.Name] == nil {
				(*i)[in.Name] = make(map[string]map[int]map[string]map[string]Param)
			}

			if (*i)[in.Name][method.Name] == nil {
				(*i)[in.Name][method.Name] = make(map[int]map[string]map[string]Param)
			}

			if (*i)[in.Name][method.Name][method.Version] == nil {
				(*i)[in.Name][method.Name][method.Version] = make(map[string]map[string]Param)
			}

			if (*i)[in.Name][method.Name][method.Version][method.HTTPmethod] == nil {
				(*i)[in.Name][method.Name][method.Version][method.HTTPmethod] = make(map[string]Param)
			}

			(*i)[in.Name][method.Name][method.Version][method.HTTPmethod][param.Name] = Param{
				Type:        param.Type,
				Optional:    param.Optional,
				Description: param.Description,
			}
		}
	}
}

type Interface map[string]map[int]map[string]map[string]Param

type Param struct {
	Type        string `json:"type"`
	Optional    bool   `json:"optional"`
	Description string `json:"description"`
}
