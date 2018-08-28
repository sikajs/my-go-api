package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

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
	var author model.User

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

	decoded := context.Get(r, "decoded")
	mapstructure.Decode(decoded.(jwt.MapClaims), &author)
	if err := dbConn.Where("username = ?", author.Username).First(&author).Error; err != nil {
		panic(err)
	} else {
		post.AuthorID = author.ID
	}

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
	var ok bool

	params := mux.Vars(r)
	id := params["id"]

	if err := json.Unmarshal([]byte(params["post"]), &v); err != nil {
		panic(err)
	}

	dbConn := db.GormConn()
	defer dbConn.Close()

	if err := dbConn.First(&post, id).Error; err != nil {
		json.NewEncoder(w).Encode("Something wrong in updating post.")
	} else {
		post.Title, ok = v["title"].(string)
		if !ok {
			fmt.Println("It's not ok to get title")
		}
		post.Content, ok = v["content"].(string)
		if !ok {
			fmt.Println("It's not ok to get content")
		}
		dbConn.Save(&post)
		httpOKAndMetaHeader(w)
		json.NewEncoder(w).Encode("Post updated.")
	}
}

//DeletePost delete a post based on id
func DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	dbConn := db.GormConn()
	defer dbConn.Close()

	if err := dbConn.Delete(&model.Post{}, id).Error; err != nil {
		panic(err)
	} else {
		httpOKAndMetaHeader(w)
		json.NewEncoder(w).Encode("Post deleted.")
	}
}
