package main

import (
	"encoding/json"
	"github.com/josestg/build-your-own-http-router/httpmux"
	"log"
	"net/http"
)

type Response struct {
	Name   string       `json:"name"`
	Method string       `json:"method"`
	Path   string       `json:"path"`
	Vars   httpmux.Vars `json:"vars"`
}

func main() {
	router := httpmux.NewRouter()

	router.HandleFunc(http.MethodGet, "/v1/users", createHandler("get users"))
	router.HandleFunc(http.MethodPost, "/v1/users", createHandler("create new user"))
	router.HandleFunc(http.MethodGet, "/v1/users/{uid}", createHandler("get users detail"))

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalln(err)
	}
}

func createHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(&Response{
			Name:   name,
			Method: r.Method,
			Path:   r.URL.Path,
			Vars:   httpmux.GetVars(r.Context()),
		})
	}
}
