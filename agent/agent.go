package agent

import (
	"log"
	"net/http"

	"github.com/brontomc/bronto/agent/api"
)

func Start() {
	router := api.SetupRouter()

	port := ":17771"
	log.Printf("Server starting on %s", port[1:])
	log.Fatal(http.ListenAndServe(port, router))
}
