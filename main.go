package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file:", err.Error())
	}

	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.Handle("/", templ.Handler(PageHome()))
	http.HandleFunc("/welcome", welcomeHandler)
	http.HandleFunc("/listen", listenHandler)
	http.Handle("/choice", templ.Handler(YesNo("speak")))
	http.HandleFunc("/speak", speakHandler)
	http.HandleFunc("/submit", submitHandler)

	fmt.Println("Listening on :4321")
	http.ListenAndServe(":4321", nil)
}