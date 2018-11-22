package web

import (
	"io/ioutil"
	"net/http"

	"github.com/gamedb/website/logging"
	"github.com/spf13/viper"
)

func rootFileHandler(w http.ResponseWriter, r *http.Request) {

	data, err := ioutil.ReadFile(viper.GetString("PATH") + r.URL.Path)
	if err != nil {
		logging.Error(err)
		_, err := w.Write([]byte("Unable to read file."))
		logging.Error(err)
		return
	}

	_, err = w.Write(data)
}
