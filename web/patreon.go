package web

import (
	"net/http"

	"github.com/go-chi/chi"
)

func patreonRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/webhooks", webhookHandler)
	return r
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {

}

func membersCreate(w http.ResponseWriter, r *http.Request) {

}

func membersUpdate(w http.ResponseWriter, r *http.Request) {

}

func membersDelete(w http.ResponseWriter, r *http.Request) {

}

func pledgeCreate(w http.ResponseWriter, r *http.Request) {

}

func pledgeUpdate(w http.ResponseWriter, r *http.Request) {

}

func pledgeDelete(w http.ResponseWriter, r *http.Request) {

}
