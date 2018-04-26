package recaptcha

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var privateKey string

var ErrEmptySecret = errors.New("empty secret")
var ErrEmptyResponse = errors.New("empty response")

func SetPrivateKey(key string) {
	privateKey = key
}

func Check(secret string, response string, ip string) (sucess bool, err error) {

	form := url.Values{}
	form.Add("secret", secret)
	form.Add("response", response)
	form.Add("remoteip", ip)

	if secret == "" {
		return false, ErrEmptySecret
	}

	if response == "" {
		return false, ErrEmptyResponse
	}

	req, err := http.NewRequest("POST", "https://www.google.com/recaptcha/api/siteverify", bytes.NewBufferString(form.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var responseStruct recaptchaResponse
	err = json.Unmarshal(responseBytes, &responseStruct)
	if err != nil {
		return false, err
	}

	return responseStruct.Success, nil
}

func CheckFromRequest(r *http.Request) (success bool, err error) {

	// Form validation
	if err := r.ParseForm(); err != nil {
		return false, err
	}

	response := r.PostForm.Get("g-recaptcha-response")
	if response == "" {
		return false, ErrEmptyResponse
	}

	success, err = Check(privateKey, response, r.RemoteAddr)
	if err != nil {
		return false, err
	}

	return success, nil
}

type recaptchaResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}
