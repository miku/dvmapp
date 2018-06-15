//go:generate statik -src=../../public
package main

import (
	"html/template"
	"log"
	"net/http"

	_ "github.com/miku/dvmapp/cmd/dvmapp/statik"
	"github.com/rakyll/statik/fs"
)

func main() {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("media"))
	http.Handle("/media/", http.StripPrefix("/media/", fs))

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(statikFS)))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("templates/index.html")
		if t == nil {
			log.Fatal("template failed")
		}
		if err != nil {
			log.Fatal(err)
		}
		if err := t.Execute(w, nil); err != nil {
			log.Fatal(err)
		}
	})
	http.ListenAndServe(":8080", nil)
}
