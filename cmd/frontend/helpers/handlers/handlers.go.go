package handlers

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gobuffalo/packr/v2"
)

func RedirectHandler(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		u, _ := url.Parse(path)
		q := u.Query()

		for k, v := range r.URL.Query() {
			for _, vv := range v {
				q.Add(k, vv)
			}
		}

		u.RawQuery = q.Encode()

		http.Redirect(w, r, u.String(), http.StatusFound)
	}
}

func RedirectHandlerFunc(f func(w http.ResponseWriter, r *http.Request) string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		u, _ := url.Parse(f(w, r))
		q := u.Query()

		for k, v := range r.URL.Query() {
			for _, vv := range v {
				q.Add(k, vv)
			}
		}

		u.RawQuery = q.Encode()

		http.Redirect(w, r, u.String(), http.StatusFound)
	}
}

func RootFileHandler(box *packr.Box, path string) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("X-Content-Type-Options", "nosniff")

		b, err := box.Find(strings.TrimPrefix(r.URL.Path, path))
		if err != nil {
			w.WriteHeader(404)
			_, err := w.Write([]byte("Unable to read file."))
			if err != nil {
				log.ErrS(err)
			}
			return
		}

		types := map[string]string{
			".js":  "text/javascript",
			".css": "text/css",
			".png": "image/png",
			".jpg": "image/jpeg",
		}

		if val, ok := types[filepath.Ext(r.URL.Path)]; ok {

			// Cache for ages
			duration := time.Hour * 1000
			w.Header().Set("Cache-Control", "max-age="+strconv.Itoa(int(duration.Seconds())))
			w.Header().Set("Expires", time.Now().Add(duration).Format(time.RFC1123))

			// Fix headers, packr seems to break them
			w.Header().Add("Content-Type", val)
		}

		// Output
		_, err = w.Write(b)
		if err != nil {
			log.ErrS(err)
		}
	}
}
