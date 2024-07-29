package api

import (
	"net/http"
)

func ExampleHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello API!"))
}
