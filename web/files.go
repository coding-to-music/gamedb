package web

import (
	"io/ioutil"
	"net/http"

	"github.com/gamedb/website/log"
	"github.com/spf13/viper"
)

func rootFileHandler(w http.ResponseWriter, r *http.Request) {

	data, err := ioutil.ReadFile(viper.GetString("PATH") + "/assets/files" + r.URL.Path)

	if err != nil {
		log.Log(err)
		w.WriteHeader(404)
		_, err := w.Write([]byte("Unable to read file."))
		log.Log(err)
		return
	}

	_, err = w.Write(data)
	log.Log(err)
}
