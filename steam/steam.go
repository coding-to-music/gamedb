package steam

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

const (
	ErrInvalidJson = "invalid character '<' looking for beginning of value"
)

func get(path string, query url.Values) (bytes []byte, err error) {

	query.Add("format", "json")
	query.Add("key", os.Getenv("STEAM_API_KEY"))

	path = "http://api.steampowered.com/" + path

	// Grab the JSON from node
	response, err := http.Get(path + "?" + query.Encode())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Convert to bytes
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// todo, test this works
	if string(contents) == "<html><head><title>Forbidden</title></head><body><h1>Forbidden</h1>Access is denied. Retrying will not help. Please verify your <pre>key=</pre> parameter.</body></html>" {
		errors.New("invalid api key")
	}

	return contents, nil
}
