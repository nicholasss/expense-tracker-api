package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("request processing...\n")
	fmt.Fprintln(w, "hello")
}

func main() {
	// setup new mux
	mux := http.NewServeMux()

	// root handler
	mux.HandleFunc("/", rootHandler)

	log.Printf("Starting server...\n")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
