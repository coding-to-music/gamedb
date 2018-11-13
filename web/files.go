package web

import (
	"io/ioutil"
	"net/http"

	"github.com/gamedb/website/logging"
	"github.com/spf13/viper"
)

func RootFileHandler(w http.ResponseWriter, r *http.Request) {

	data, err := ioutil.ReadFile(viper.GetString("PATH") + r.URL.Path)
	if err != nil {
		logging.Error(err)
		w.Write([]byte("Unable to read file."))
		return
	}

	w.Write(data)
}
