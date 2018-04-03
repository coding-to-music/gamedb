package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/rollbar/rollbar-go"
)

const (
	prod = "production"
)

func Error(err error) {

	if err != nil {

		fmt.Println(time.Now().Format(time.Stamp) + ": " + err.Error())

		if os.Getenv("ENV") == prod {
			rollbar.Error(rollbar.ERR, err)
		}
	}
}

func Info(message string) {

	fmt.Println(time.Now().Format(time.Stamp) + ": " + message)

	if os.Getenv("ENV") == prod {
		rollbar.Message(rollbar.INFO, message)
	}
}
