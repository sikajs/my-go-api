package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

const (
	host      = "localhost"
	port      = 5432
	user      = "pqgotest"
	password  = "password"
	dbname    = "pqgotest"
	sslmode   = "disable"
	secretKey = "Welcome to JS's playground"
)

// Post data structure
type Post struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	AuthorID int    `json:"auther_id"`
}

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
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
	router.HandleFunc("/get_token/{user}", getToken).Methods("POST")
	// router.HandleFunc("/test", validateMiddleware(TestEndPoint)).Methods("GET")

	routesV1(router).HandleFunc("/posts/{post}", validateMiddleware(createPost)).Methods("POST")
	routesV1(router).HandleFunc("/posts/", listPosts).Methods("GET")
	routesV1(router).HandleFunc("/posts/{id}", showPost).Methods("GET")
	routesV1(router).HandleFunc("/posts/{id}/{post}", validateMiddleware(updatePost)).Methods("PUT")
	routesV1(router).HandleFunc("/posts/{id}", validateMiddleware(deletePost)).Methods("DELETE")

	fmt.Println("Running server!")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("Still alive!")
}

func getToken(w http.ResponseWriter, r *http.Request) {
	var v map[string]interface{}
	var user User
	var ok bool
	var passwordParam string
	var finalPass string

	params := mux.Vars(r)
	if err := json.Unmarshal([]byte(params["user"]), &v); err != nil {
		panic(err)
	}
	user.Username, ok = v["username"].(string)
	if !ok {
		fmt.Println("It's not ok to get username")
	}
	passwordParam, ok = v["password"].(string)
	if !ok {
		fmt.Println("It's not ok to get password")
	}

	db := dbConnect()
	defer db.Close()

	selectUserSQL := `SELECT id, name, password FROM users WHERE username=$1`
	result := db.QueryRow(selectUserSQL, user.Username)
	switch err := result.Scan(&user.ID, &user.Name, &user.Password); err {
	case sql.ErrNoRows:
		fmt.Println("No row was returned.")
	case nil:
		encryptedPass := md5.Sum([]byte(passwordParam))
		finalPass = hex.EncodeToString(encryptedPass[:])

		if user.Password != finalPass {
			fmt.Println("Invalid login requested.")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Authentication failed"))
			return
		} else {
			claims := make(jwt.MapClaims)
			claims["exp"] = time.Now().Add(time.Hour * time.Duration(1)).Unix()
			claims["iat"] = time.Now().Unix()
			claims["id"] = user.ID
			claims["username"] = user.Username
			claims["password"] = user.Password

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString([]byte(secretKey))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "Error while signing the token")
				log.Fatal(err)
			}

			// json.NewEncoder(w).Encode(tokenString)
			w.Write([]byte(tokenString))
		}
	default:
		panic(err)
	}
}

// func TestEndPoint(w http.ResponseWriter, r *http.Request) {
// 	decoded := context.Get(r, "decoded")
// 	var user User
// 	mapstructure.Decode(decoded.(jwt.MapClaims), &user)
// 	json.NewEncoder(w).Encode(user)
// }

func validateMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("authorization")
		if authorizationHeader != "" {
			bearerToken := strings.Split(authorizationHeader, " ")
			if len(bearerToken) == 2 {
				token, error := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("There was an error")
					}
					return []byte(secretKey), nil
				})
				if error != nil {
					json.NewEncoder(w).Encode(error.Error())
					return
				}
				if token.Valid {
					log.Println("TOKEN WAS VALID")
					context.Set(r, "decoded", token.Claims)
					next(w, r)
				} else {
					json.NewEncoder(w).Encode("Invalid authorization token")
				}
			}
		} else {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode("An authorization header is required.")
		}
	})
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
	var user User

	decoded := context.Get(r, "decoded")
	mapstructure.Decode(decoded.(jwt.MapClaims), &user)
	post.AuthorID = user.ID

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

	createPostSQL := `
	INSERT INTO posts (id, title, content, author_id)
	VALUES ($1, $2, $3, $4)
	RETURNING id`

	dbErr := db.QueryRow(createPostSQL, post.ID, post.Title, post.Content, post.AuthorID).Scan(&id)
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

	selectPostsSQL := `SELECT id, title, content, author_id From posts ORDER BY id`
	rows, dbErr := db.Query(selectPostsSQL)
	if dbErr != nil {
		panic(dbErr)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID)
		if err != nil {
			panic(err)
		}
		posts = append(posts, Post{post.ID, post.Title, post.Content, post.AuthorID})
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

	selectPostSQL := `SELECT id, title, content, author_id FROM posts WHERE id=$1`
	row := db.QueryRow(selectPostSQL, key)
	switch err := row.Scan(&p.ID, &p.Title, &p.Content, &p.AuthorID); err {
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

	delSQL := `DELETE FROM posts WHERE id=$1`
	switch _, err := db.Exec(delSQL, key); err {
	case nil:
		httpOKAndMetaHeader(w)
		json.NewEncoder(w).Encode("Post deleted.")
	default:
		panic(err)
	}
}
