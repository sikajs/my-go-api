package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	"github.com/sikajs/my-go-api/route"
)

func main() {
	loggedRouter := serverLoggingHandler(route.Routes())
	fmt.Println("Running server!")
	http.ListenAndServe(decidePort(), loggedRouter)
}

func decidePort() string {
	port := os.Getenv("PORT")
	return ":" + port
}

//serverLoggingHandler
func serverLoggingHandler(h http.Handler) http.Handler {
	logFile, err := os.OpenFile("server.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	return handlers.LoggingHandler(logFile, h)
}

// parameter filter and check, according to the data model, filter out the one not belongs to the model?
// func parameterMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		next.ServeHTTP(w, r)
// 	})
// }
