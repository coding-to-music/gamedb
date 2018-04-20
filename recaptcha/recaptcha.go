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

var ErrInvalidInputs = errors.New("invalid inputs")

func Check(secret string, response string, ip string) (sucess bool, err error) {

	form := url.Values{}
	form.Add("secret", secret)
	form.Add("response", response)
	form.Add("remoteip", ip)

	if secret == "" || response == "" {
		return false, ErrInvalidInputs
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

type recaptchaResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}
