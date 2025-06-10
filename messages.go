package main

import (
	"net/http"
	"time"
)

func submitHandler(w http.ResponseWriter, req *http.Request) {
	// Htmx requests send URL-encoded data
	err := req.ParseForm()
    if err != nil {
        http.Error(w, "Error reading body", http.StatusBadRequest)
        return
    }

	time.Sleep(2 * time.Second)
    
	// TODO: profanity filter
	// TODO: input sanitation
	// TODO: write response to csv in s3

    err = UserMessage(req.FormValue("userInput")).Render(req.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusBadRequest)
		return
    }

    err = NyxMessage("Nice message!").Render(req.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusBadRequest)
		return
    }

	println("User sent:", req.FormValue("userInput"))
}