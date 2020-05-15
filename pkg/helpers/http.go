package helpers

import (
	"net/http"
	"time"
)

func GetWithTimeout(url string, timeout time.Duration) (*http.Response, error) {

	if timeout == 0 {
		timeout = time.Second * 10
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	clientWithTimeout := &http.Client{
		Timeout: timeout,
	}

	return clientWithTimeout.Do(req)
}
