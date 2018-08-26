package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/sikajs/my-go-api/db"
	"github.com/sikajs/my-go-api/model"
)

func httpOKAndMetaHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

//CreatePost creates a post from parameters
func CreatePost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var post model.Post
	var v map[string]interface{}
	var ok bool
	var user model.User

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

	dbConn := db.GormConn()
	defer dbConn.Close()

	if dbConn.NewRecord(post) {
		dbConn.Create(&post)
		httpOKAndMetaHeader(w)
		json.NewEncoder(w).Encode(map[string]uint{"post": post.ID})
	} else {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		json.NewEncoder(w).Encode("Error: post already existed!")
	}
}

//ListPosts lists all posts in db
func ListPosts(w http.ResponseWriter, r *http.Request) {
	var posts []model.Post

	dbConn := db.GormConn()
	defer dbConn.Close()

	if err := dbConn.Find(&posts).Error; err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		json.NewEncoder(w).Encode("Error in finding posts")
	} else {
		httpOKAndMetaHeader(w)
		json.NewEncoder(w).Encode(posts)
	}
}

//ShowPost display post detail
func ShowPost(w http.ResponseWriter, r *http.Request) {
	var p model.Post
	vars := mux.Vars(r)
	id := vars["id"]

	dbConn := db.GormConn()
	defer dbConn.Close()

	if err := dbConn.First(&p, id).Error; err != nil {
		fmt.Println("No row were found with key ", id)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode("No row were found")
	} else {
		httpOKAndMetaHeader(w)
		json.NewEncoder(w).Encode(map[string]model.Post{"post": p})
	}
}

//UpdatePost updates a post with parameters
func UpdatePost(w http.ResponseWriter, r *http.Request) {
	var v map[string]interface{}
	var post model.Post
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

	dbConn := db.Connect()
	defer dbConn.Close()

	checkStatement := `SELECT id FROM posts WHERE id=$1`
	row := dbConn.QueryRow(checkStatement, key)
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
		if _, updateErr := dbConn.Exec(updateStatement); err != nil {
			panic(updateErr)
		}
		httpOKAndMetaHeader(w)
		json.NewEncoder(w).Encode("Post updated.")
	default:
		panic(err)
	}
}

//DeletePost delete a post based on id
func DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]

	dbConn := db.Connect()
	defer dbConn.Close()

	delSQL := `DELETE FROM posts WHERE id=$1`
	switch _, err := dbConn.Exec(delSQL, key); err {
	case nil:
		httpOKAndMetaHeader(w)
		json.NewEncoder(w).Encode("Post deleted.")
	default:
		panic(err)
	}
}
