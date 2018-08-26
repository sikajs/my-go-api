package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/lib/pq"
	"github.com/sikajs/my-go-api/route"
)

func main() {
	fmt.Println("Running server!")
	log.Fatal(http.ListenAndServe(decidePort(), route.Routes()))
}

func decidePort() string {
	port := os.Getenv("PORT")
	return ":" + port
}

// parameter filter and check, according to the data model, filter out the one not belongs to the model?
// func parameterMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		next.ServeHTTP(w, r)
// 	})
// }
