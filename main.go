package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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
