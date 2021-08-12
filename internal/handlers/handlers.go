package handlers

import (
	"log"
	"net/http"

	"github.com/CloudyKit/jet/v6"
)

var (
	views = jet.NewSet(
		jet.NewOSFileSystemLoader("html"),
		jet.InDevelopmentMode(),
	)
)

func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	if err := view.Execute(w, data, nil); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func Home(w http.ResponseWriter, r *http.Request) {
	if err := renderPage(w, "home.html", nil); err != nil {
		log.Println(err)
	}
}
