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
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// DB connection
func dbConnect() *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(`Database connected.`)
	return db
}

func main() {
	// routing
	var router = mux.NewRouter()
	router.HandleFunc("/healthcheck", healthCheck).Methods("GET")
	// router.HandleFunc("/message", handleQryMessage).Methods("GET")

	router.HandleFunc("/posts/{title}", createPost).Methods("POST")
	router.HandleFunc("/posts/", listPosts).Methods("GET")
	router.HandleFunc("/posts/{id}", showPost).Methods("GET")
	router.HandleFunc("/posts/{id}", updatePost).Methods("PUT")
	router.HandleFunc("/posts/{id}", deletePost).Methods("DELETE")

	fmt.Println("Running server!")
	log.Fatal(http.ListenAndServe(":8000", router))
}

// func handleQryMessage(w http.ResponseWriter, r *http.Request) {
// 	vars := r.URL.Query()
// 	message := vars.Get("msg")

// 	json.NewEncoder(w).Encode(map[string]string{"message": message})
// }

// func handleURLMessage(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	message := vars["msg"]

// 	json.NewEncoder(w).Encode(map[string]string{"message_via_url": message})
// }

func healthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("Still alive!")
}

// CRUD for post
func createPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var post Post
	var id int
	var currMaxID int
	var nextID int

	_ = json.NewDecoder(r.Body).Decode(&post)
	post.Title = params["title"]
	post.Content = params["content"]

	//TODO: DB connection, will re-connect db every time called.
	db := dbConnect()
	defer db.Close()

	getMaxSQL := `SELECT MAX(id) FROM posts`
	row := db.QueryRow(getMaxSQL)
	switch err := row.Scan(&currMaxID); err {
	case sql.ErrNoRows:
		fmt.Println("No row was returned!")
	case nil:
		nextID = currMaxID + 1
	default:
		panic(err)
	}

	sqlStatement := `
	INSERT INTO posts (id, title, content)
	VALUES ($1, $2, $3)
	RETURNING id`

	dbErr := db.QueryRow(sqlStatement, nextID, post.Title, post.Content).Scan(&id)
	if dbErr != nil {
		panic(dbErr)
	}
	fmt.Println("New record ID is:", id)

	json.NewEncoder(w).Encode(map[string]int{"post": id})
}

//list
func listPosts(w http.ResponseWriter, r *http.Request) {
	db := dbConnect()
	defer db.Close()

	sqlStatement := `SELECT id, title, content From posts`
	rows, dbErr := db.Query(sqlStatement)
	if dbErr != nil {
		panic(dbErr)
	}
	defer rows.Close()

	var posts []Post
	var post Post
	for rows.Next() {
		err := rows.Scan(&post.ID, &post.Title, &post.Content)
		if err != nil {
			panic(err)
		}
		posts = append(posts, Post{post.ID, post.Title, post.Content})
	}

	json.NewEncoder(w).Encode(posts)
}

//show
func showPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]

	db := dbConnect()
	defer db.Close()

	var p Post
	sqlStatement := `SELECT id, title, content FROM posts WHERE id=$1`
	row := db.QueryRow(sqlStatement, key)
	switch err := row.Scan(&p.ID, &p.Title, &p.Content); err {
	case sql.ErrNoRows:
		fmt.Println("No row were found with key ", key)
	case nil:
		json.NewEncoder(w).Encode(map[string]Post{"post": p})
	default:
		panic(err)
	}
}

func updatePost(w http.ResponseWriter, r *http.Request) {
}

func deletePost(w http.ResponseWriter, r *http.Request) {
}
