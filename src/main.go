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

const (
	host     = "localhost"
	port     = 5432
	user     = "pqgotest"
	password = "password"
	dbname   = "pqgotest"
	sslmode  = "disable"
)

// Post data structure
type Post struct {
	id      int
	title   string
	content string
}

func main() {
	// routing
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

// DB connection
func dbConnect() {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println(`Database connected.`)
}

// CRUD for post
func createPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	message := vars["msg"]

	//DB connection
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println(`Database connected.`)

	sqlStatement := `
	INSERT INTO posts (id, title, content)
	VALUES ($1, $2, $3)
	RETURNING id`

	var id int

	dbErr := db.QueryRow(sqlStatement, 5, "insert from go", "nice description").Scan(&id)
	if dbErr != nil {
		panic(dbErr)
	}
	fmt.Println("New record ID is:", id)

	json.NewEncoder(w).Encode(map[string]string{"create post based on ": message})
}

func listPosts(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// message := vars[""]

	json.NewEncoder(w).Encode(map[string]string{"posts": "list all the posts"})
}

func showPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	var p Post

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println(`Database connected.`)

	sqlStatement := `SELECT id, title, content FROM posts WHERE id=$1`
	row := db.QueryRow(sqlStatement, key)
	switch err := row.Scan(&p.id, &p.title, &p.content); err {
	case sql.ErrNoRows:
		fmt.Println("No row were found with key ", key)
	case nil:
		fmt.Println(p)
		json.NewEncoder(w).Encode(map[string]string{"post": p.title})
	default:
		panic(err)
	}

}
