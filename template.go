package main

import (
	"html/template"
	"net/http"
	"path"
	"runtime"
	"strings"

	"github.com/Jleagle/go-helpers/logger"
)

func returnTemplate(w http.ResponseWriter, page string, pageData interface{}) {

	// Get current app path
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		logger.Info("Failed to get path")
	}
	folder := path.Dir(file)

	// Load templates needed
	t, err := template.ParseFiles(folder+"/templates/header.html", folder+"/templates/footer.html", folder+"/templates/"+page+".html")
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			returnErrorTemplate(w, 404, "The file for this page seems to be missing!")
			return
		}
		logger.Error(err)
	}

	// Write a respone
	err = t.ExecuteTemplate(w, page, pageData)
	if err != nil {
		logger.Error(err)
	}
}

func returnErrorTemplate(w http.ResponseWriter, code int, message string) {

	template := errorTemplate{
		Code:    code,
		Message: message,
	}

	returnTemplate(w, "error", template)
}

type errorTemplate struct {
	Code    int
	Message string
}
