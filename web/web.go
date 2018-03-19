package web

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/go-helpers/logger"
	"github.com/dustin/go-humanize"
	"github.com/gosimple/slug"
	"github.com/kr/pretty"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/session"
	"github.com/steam-authority/steam-authority/steam"
)

func returnTemplate(w http.ResponseWriter, r *http.Request, page string, pageData interface{}) (err error) {

	// Load templates needed
	folder := os.Getenv("STEAM_PATH")
	t, err := template.New("t").Funcs(getTemplateFuncMap()).ParseFiles(folder+"/templates/_header.html", folder+"/templates/_footer.html", folder+"/templates/"+page+".html")
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 404, err.Error())
		return err
	}

	// Write a respone
	buf := &bytes.Buffer{}
	err = t.ExecuteTemplate(buf, page, pageData)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, "Something has gone wrong, the error has been logged!")
		return
	} else {
		// No error, send the content, HTTP 200 response status implied
		buf.WriteTo(w)
		return
	}

	return nil
}

func returnErrorTemplate(w http.ResponseWriter, r *http.Request, code int, message string) {

	w.WriteHeader(code)

	tmpl := errorTemplate{}
	tmpl.Fill(r, "Error")
	tmpl.Code = code
	tmpl.Message = message

	returnTemplate(w, r, "error", tmpl)
}

type errorTemplate struct {
	GlobalTemplate
	Code    int
	Message string
}

func getTemplateFuncMap() map[string]interface{} {
	return template.FuncMap{
		"join":   func(a []string) string { return strings.Join(a, ", ") },
		"title":  func(a string) string { return strings.Title(a) },
		"comma":  func(a int) string { return humanize.Comma(int64(a)) },
		"commaf": func(a float64) string { return humanize.Commaf(a) },
		"slug":   func(a string) string { return slug.Make(a) },
		"apps": func(a []int, appsMap map[int]mysql.App) template.HTML {
			var apps []string
			for _, v := range a {
				apps = append(apps, "<a href=\"/apps/"+strconv.Itoa(v)+"\">"+appsMap[v].GetName()+"</a>")
			}
			return template.HTML("Apps: " + strings.Join(apps, ", "))
		},
		"packages": func(a []int, packagesMap map[int]mysql.Package) template.HTML {
			var packages []string
			for _, v := range a {
				packages = append(packages, "<a href=\"/packages/"+strconv.Itoa(v)+"\">"+packagesMap[v].GetName()+"</a>")
			}
			return template.HTML("Packages: " + strings.Join(packages, ", "))
		},
		"tags": func(a []mysql.Tag) template.HTML {
			var tags []string
			for _, v := range a {
				tags = append(tags, "<a href=\"/apps?tags="+strconv.Itoa(v.ID)+"\">"+v.GetName()+"</a>")
			}
			return template.HTML(strings.Join(tags, ", "))
		},
		"genres": func(a []steam.AppDetailsGenre) template.HTML {
			var genres []string
			for _, v := range a {
				genres = append(genres, "<a href=\"/apps?genres="+strconv.Itoa(v.ID)+"\">"+v.Description+"</a>")
			}
			return template.HTML(strings.Join(genres, ", "))
		},
		"unix":       func(t time.Time) int64 { return t.Unix() },
		"startsWith": func(a string, b string) bool { return strings.HasPrefix(a, b) },
		"contains":   func(a string, b string) bool { return strings.Contains(a, b) },
	}
}

// GlobalTemplate is added to every other template
type GlobalTemplate struct {
	Env     string
	ID      int
	Name    string // Username
	Title   string // Page title
	Avatar  string
	Level   int
	Games   []int
	Path    string // URL
	IsAdmin bool
	request *http.Request // Internal
}

func (t *GlobalTemplate) Fill(r *http.Request, title string) {

	t.Title = title
	t.request = r

	// From ENV
	t.Env = os.Getenv("ENV")

	// From session
	id, _ := session.Read(r, session.ID)
	level, _ := session.Read(r, session.Level)

	t.ID, _ = strconv.Atoi(id)
	t.Name, _ = session.Read(r, session.Name)
	t.Avatar, _ = session.Read(r, session.Avatar)
	t.Avatar, _ = session.Read(r, session.Avatar)
	t.Level, _ = strconv.Atoi(level)

	gamesString, _ := session.Read(r, session.Games)
	if gamesString != "" {
		err := json.Unmarshal([]byte(gamesString), &t.Games)
		if err != nil {
			logger.Error(err)
			if strings.Contains(err.Error(), "cannot unmarshal") {
				pretty.Print(gamesString)
			}
		}
	}

	// From request
	t.Path = r.URL.Path
	t.IsAdmin = r.Header.Get("Authorization") != ""
}

func (t GlobalTemplate) LoggedIn() (bool) {
	return t.ID > 0
}

func (t GlobalTemplate) IsLocal() (bool) {
	return t.Env == "local"
}

func (t GlobalTemplate) IsProduction() (bool) {
	return t.Env == "production"
}

func (t GlobalTemplate) ShowAd() (bool) {

	noAds := []string{
		"/admin",
		"/donate",
	}

	for _, v := range noAds {
		if strings.HasPrefix(t.request.URL.Path, v) {
			return false
		}
	}

	return true
}
