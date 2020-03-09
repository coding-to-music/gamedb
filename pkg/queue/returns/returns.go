package returns

import (
	"github.com/Jleagle/rabbit-go"
)

type Action int

const (
	ActionFail Action = iota
	ActionRetry
	ActionSuccess
)

// type returnInterface interface {
// 	action() string
// 	message() string
// }

type Response struct {
	action  Action
	message *rabbit.Message
}

func SuccessResponse(msg *rabbit.Message) Response {
	return Response{
		action:  ActionSuccess,
		message: msg,
	}
}

func RetryResponse(msg *rabbit.Message) Response {
	return Response{
		action:  ActionRetry,
		message: msg,
	}
}

func FailResponse(msg *rabbit.Message) Response {
	return Response{
		action:  ActionFail,
		message: msg,
	}
}
