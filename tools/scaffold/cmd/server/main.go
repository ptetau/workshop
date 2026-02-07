package main

import (
	"log"
	"net/http"

	web "workshop/internal/adapters/http"
)

func main() {
	mux := web.NewMux("./static")

	addr := ":8080"
	server := &http.Server{Addr: addr, Handler: mux}

	log.Printf("listening on %s", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
