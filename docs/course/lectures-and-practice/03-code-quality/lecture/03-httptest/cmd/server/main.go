package main

import (
	"log"
	"net/http"

	httptestdemo "code-quality-examples/03-httptest"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/greet", httptestdemo.GreetingHandler)

	const address = ":8080"
	log.Printf("demo server is listening on http://localhost%s/greet?name=Alice", address)
	if err := http.ListenAndServe(address, mux); err != nil {
		log.Fatal(err)
	}
}
