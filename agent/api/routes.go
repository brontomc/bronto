package api

import (
	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", ExampleHandler).Methods("GET")

	return r
}
