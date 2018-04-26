package recaptcha

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const Response = "g-recaptcha-response"

var privateKey string

var ErrEmptySecret = Error{"empty secret", false}
var ErrEmptyResponse = Error{"empty response", false}
var ErrNotChecked = Error{"captcha not checked", true}

func SetPrivateKey(key string) {
	privateKey = key
}

func Check(secret string, response string, ip string) (err error) {

	form := url.Values{}
	form.Add("secret", secret)
	form.Add("response", response)
	form.Add("remoteip", ip)

	if secret == "" {
		return ErrEmptySecret
	}

	if response == "" {
		return ErrEmptyResponse
	}

	req, err := http.NewRequest("POST", "https://www.google.com/recaptcha/api/siteverify", bytes.NewBufferString(form.Encode()))
	if err != nil {
		return Error{err.Error(), false}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Error{err.Error(), false}
	}
	defer resp.Body.Close()

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Error{err.Error(), false}
	}

	var responseStruct recaptchaResponse
	err = json.Unmarshal(responseBytes, &responseStruct)
	if err != nil {
		return Error{err.Error(), false}
	}

	if !responseStruct.Success {
		return ErrNotChecked
	}

	return nil
}

func CheckFromRequest(r *http.Request) (err error) {

	// Form validation
	if err := r.ParseForm(); err != nil {
		return Error{err.Error(), false}
	}

	response := r.PostForm.Get(Response)
	if response == "" {
		return ErrNotChecked
	}

	return Check(privateKey, response, r.RemoteAddr)
}

type recaptchaResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

type Error struct {
	error     string
	userError bool
}

func (r Error) Error() string {
	return r.error
}

func (r Error) IsUserError() bool {
	return r.userError
}
