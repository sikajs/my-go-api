package route

import (
	"github.com/gorilla/mux"
	"github.com/sikajs/my-go-api/handler"
)

//Routes returns all of the setted routes
func Routes() *mux.Router {
	// routing
	var router = mux.NewRouter()
	router.HandleFunc("/healthcheck", handler.HealthCheck).Methods("GET")
	router.HandleFunc("/get_token/{user}", handler.GetToken).Methods("POST")

	routesV1(router).HandleFunc("/posts/{post}", handler.ValidateMiddleware(handler.CreatePost)).Methods("POST")
	routesV1(router).HandleFunc("/posts/", handler.ListPosts).Methods("GET")
	routesV1(router).HandleFunc("/posts/{id}", handler.ShowPost).Methods("GET")
	routesV1(router).HandleFunc("/posts/{id}/{post}", handler.ValidateMiddleware(handler.UpdatePost)).Methods("PUT")
	routesV1(router).HandleFunc("/posts/{id}", handler.ValidateMiddleware(handler.DeletePost)).Methods("DELETE")

	return router
}

func routesV1(r *mux.Router) *mux.Router {
	return r.PathPrefix("/v1").Subrouter()
}
