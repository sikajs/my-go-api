package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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

func routesV1(r *mux.Router) *mux.Router {
	return r.PathPrefix("/v1").Subrouter()
}

func httpOKAndMetaHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func main() {
	// routing
	var router = mux.NewRouter()
	router.HandleFunc("/healthcheck", healthCheck).Methods("GET")

	routesV1(router).HandleFunc("/posts/{post}", createPost).Methods("POST")
	routesV1(router).HandleFunc("/posts/", listPosts).Methods("GET")
	routesV1(router).HandleFunc("/posts/{id}", showPost).Methods("GET")
	routesV1(router).HandleFunc("/posts/{id}/{post}", updatePost).Methods("PUT")
	routesV1(router).HandleFunc("/posts/{id}", deletePost).Methods("DELETE")

	fmt.Println("Running server!")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("Still alive!")
}

// CRUD for post
//create
func createPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var post Post
	var id int
	var currMaxID int
	var v map[string]interface{}
	var ok bool

	if err := json.Unmarshal([]byte(params["post"]), &v); err != nil {
		panic(err)
	}
	post.Title, ok = v["title"].(string)
	if !ok {
		fmt.Println("It's not ok to get title")
	}
	post.Content, ok = v["content"].(string)
	if !ok {
		fmt.Println("It's not ok to get content")
	}

	// TODO: DB connection, will re-connect db every time called.
	db := dbConnect()
	defer db.Close()

	// TODO: need to find better way to get next id
	getMaxSQL := `SELECT MAX(id) FROM posts`
	row := db.QueryRow(getMaxSQL)
	switch err := row.Scan(&currMaxID); err {
	case sql.ErrNoRows:
		fmt.Println("No row was returned!")
	case nil:
		post.ID = currMaxID + 1
	default:
		panic(err)
	}

	sqlStatement := `
	INSERT INTO posts (id, title, content)
	VALUES ($1, $2, $3)
	RETURNING id`

	dbErr := db.QueryRow(sqlStatement, post.ID, post.Title, post.Content).Scan(&id)
	if dbErr != nil {
		panic(dbErr)
	}
	fmt.Println("New record ID is:", id)

	httpOKAndMetaHeader(w)
	json.NewEncoder(w).Encode(map[string]int{"post": id})
}

//list
func listPosts(w http.ResponseWriter, r *http.Request) {
	var posts []Post
	var post Post

	db := dbConnect()
	defer db.Close()

	sqlStatement := `SELECT id, title, content From posts`
	rows, dbErr := db.Query(sqlStatement)
	if dbErr != nil {
		panic(dbErr)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&post.ID, &post.Title, &post.Content)
		if err != nil {
			panic(err)
		}
		posts = append(posts, Post{post.ID, post.Title, post.Content})
	}

	httpOKAndMetaHeader(w)
	json.NewEncoder(w).Encode(posts)
}

//show
func showPost(w http.ResponseWriter, r *http.Request) {
	var p Post

	vars := mux.Vars(r)
	key := vars["id"]

	db := dbConnect()
	defer db.Close()

	sqlStatement := `SELECT id, title, content FROM posts WHERE id=$1`
	row := db.QueryRow(sqlStatement, key)
	switch err := row.Scan(&p.ID, &p.Title, &p.Content); err {
	case sql.ErrNoRows:
		fmt.Println("No row were found with key ", key)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("No row were found")
	case nil:
		httpOKAndMetaHeader(w)
		json.NewEncoder(w).Encode(map[string]Post{"post": p})
	default:
		panic(err)
	}
}

//update
func updatePost(w http.ResponseWriter, r *http.Request) {
	var v map[string]interface{}
	var post Post
	var hasCondition bool
	var ok bool

	params := mux.Vars(r)
	key := params["id"]

	if err := json.Unmarshal([]byte(params["post"]), &v); err != nil {
		panic(err)
	}
	post.Title, ok = v["title"].(string)
	if !ok {
		fmt.Println("It's not ok to get title")
	}
	post.Content, ok = v["content"].(string)
	if !ok {
		fmt.Println("It's not ok to get content")
	}

	db := dbConnect()
	defer db.Close()

	checkStatement := `SELECT id FROM posts WHERE id=$1`
	row := db.QueryRow(checkStatement, key)
	switch err := row.Scan(&post.ID); err {
	case sql.ErrNoRows:
		fmt.Println("No row were found with key ", key)
		httpOKAndMetaHeader(w)
		json.NewEncoder(w).Encode("No row were found")
	case nil:
		updateStatement := `UPDATE posts SET `
		if post.Title != "" {
			updateStatement = strings.Join([]string{updateStatement, " title='", post.Title, "'"}, "")
			hasCondition = true
		}
		if post.Content != "" {
			if hasCondition {
				updateStatement = strings.Join([]string{updateStatement, ","}, "")
			}
			updateStatement = strings.Join([]string{updateStatement, " content='", post.Content, "'"}, "")
		}
		updateStatement = strings.Join([]string{updateStatement, " WHERE id='", key, "'"}, "")
		if _, updateErr := db.Exec(updateStatement); err != nil {
			panic(updateErr)
		}
		httpOKAndMetaHeader(w)
		json.NewEncoder(w).Encode("Post updated.")
	default:
		panic(err)
	}
}

//delete
func deletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]

	db := dbConnect()
	defer db.Close()

	sqlStatement := `DELETE FROM posts WHERE id=$1`
	switch _, err := db.Exec(sqlStatement, key); err {
	case nil:
		httpOKAndMetaHeader(w)
		json.NewEncoder(w).Encode("Post deleted.")
	default:
		panic(err)
	}
}
