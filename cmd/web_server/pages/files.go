package pages

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func rootFileHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*24)

	w.Header().Set("X-Content-Type-Options", "nosniff")

	r.URL.Path = strings.Replace(r.URL.Path, "..", "", -1)
	data, err := ioutil.ReadFile(config.Config.GameDBDirectory.Get() + "/assets/files" + r.URL.Path)

	if err != nil {
		log.Err(err, r)
		w.WriteHeader(404)
		_, err := w.Write([]byte("Unable to read file."))
		log.Err(err, r)
		return
	}

	_, err = w.Write(data)
	log.Err(err, r)
}
