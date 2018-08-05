package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
)

func main() {
	//DB connection
	connStr := "user=pqgotest dbname=pqgotest"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//DB query
	var id int
	err := db.QueryRow("INSERT INTO posts(title,content) VALUES('test', 'desc123') RETURNING id").Scan(&id)
	fmt.Println(id)
	if err != nil {
		log.Fatal(err)
	}

	var router = mux.NewRouter()
	router.HandleFunc("/healthcheck", healthCheck).Methods("GET")
	router.HandleFunc("/message", handleQryMessage).Methods("GET")

	router.HandleFunc("/posts/create/{post_param}", createPost).Methods("POST")
	router.HandleFunc("/posts/", listPosts).Methods("GET")
	router.HandleFunc("/posts/{key}", showPost).Methods("GET")

	fmt.Println("Running server!")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func handleQryMessage(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	message := vars.Get("msg")

	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func handleURLMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	message := vars["msg"]

	json.NewEncoder(w).Encode(map[string]string{"message_via_url": message})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("Still alive!")
}

//

//CRUD for post
func createPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	message := vars["msg"]

	json.NewEncoder(w).Encode(map[string]string{"create post based on ": message})
}

func listPosts(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// message := vars["msg"]

	json.NewEncoder(w).Encode(map[string]string{"posts": "list all the posts"})
}

func showPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	json.NewEncoder(w).Encode(map[string]string{"list the post based on key": key})
}
