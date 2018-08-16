package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
	"github.com/sikajs/my-go-api/handler"
)

func routesV1(r *mux.Router) *mux.Router {
	return r.PathPrefix("/v1").Subrouter()
}

func main() {
	// routing
	var router = mux.NewRouter()
	router.HandleFunc("/healthcheck", healthCheck).Methods("GET")
	router.HandleFunc("/get_token/{user}", handler.GetToken).Methods("POST")

	routesV1(router).HandleFunc("/posts/{post}", handler.ValidateMiddleware(handler.CreatePost)).Methods("POST")
	routesV1(router).HandleFunc("/posts/", handler.ListPosts).Methods("GET")
	routesV1(router).HandleFunc("/posts/{id}", handler.ShowPost).Methods("GET")
	routesV1(router).HandleFunc("/posts/{id}/{post}", handler.ValidateMiddleware(handler.UpdatePost)).Methods("PUT")
	routesV1(router).HandleFunc("/posts/{id}", handler.ValidateMiddleware(handler.DeletePost)).Methods("DELETE")

	fmt.Println("Running server!")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("Still alive!")
}
