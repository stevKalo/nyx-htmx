package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
)

func main() {
	home := PageHome()
	about := PageAbout()

	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.Handle("/", templ.Handler(home))
	http.Handle("/about", templ.Handler(about))

	fmt.Println("Listening on :4321")
	http.ListenAndServe(":4321", nil)
}
