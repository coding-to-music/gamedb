package web

import (
	"net/http"

	"github.com/Jleagle/go-helpers/logger"
	"github.com/steam-authority/steam-authority/queue"
)

func QueuesHandler(w http.ResponseWriter, r *http.Request) {

	queues, err := queue.GetQeueus()
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Template
	template := queueTemplate{}
	template.Fill(r, "Queues")
	template.Queues = queues

	returnTemplate(w, r, "queues", template)
	return
}

type queueTemplate struct {
	GlobalTemplate
	Queues []queue.Queue
}
