package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

var indexTemplate = template.Must(template.ParseFiles("templates/index.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	indexTemplate.Execute(w, nil)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Println("Starting server on port " + port)

	fs := http.FileServer(http.Dir("public"))

	mux := http.NewServeMux()

	mux.Handle("/public/", http.StripPrefix("/public/", fs))
	mux.Handle("/", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(indexHandler)))

	http.ListenAndServe(":"+port, mux)
}
