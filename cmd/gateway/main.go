package main

import (
	"fmt"
	"net/http"
)

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request:", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"message": "Hello World"}`)
}

func pinghandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request:", r.URL.Path)
	fmt.Fprintln(w, "PONG")
}

func main() {
	fmt.Println("Starting API Gateway")

	http.HandleFunc("/", helloWorld)
	http.HandleFunc("/ping", pinghandler)

	http.ListenAndServe(":8080", nil)
}
