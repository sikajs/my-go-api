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
	"github.com/sikajs/my-go-api/db"
	"github.com/sikajs/my-go-api/handler"
)

const (
	secretKey = "Welcome to JS's playground"
)

// User data structure
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
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

	routesV1(router).HandleFunc("/posts/{post}", validateMiddleware(handler.CreatePost)).Methods("POST")
	routesV1(router).HandleFunc("/posts/", handler.ListPosts).Methods("GET")
	routesV1(router).HandleFunc("/posts/{id}", handler.ShowPost).Methods("GET")
	routesV1(router).HandleFunc("/posts/{id}/{post}", validateMiddleware(handler.UpdatePost)).Methods("PUT")
	routesV1(router).HandleFunc("/posts/{id}", validateMiddleware(handler.DeletePost)).Methods("DELETE")

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

	dbConn := db.Connect()
	defer dbConn.Close()

	selectUserSQL := `SELECT id, name, password FROM users WHERE username=$1`
	result := dbConn.QueryRow(selectUserSQL, user.Username)
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

// func TestEndPoint(w http.ResponseWriter, r *http.Request) {
// 	decoded := context.Get(r, "decoded")
// 	var user User
// 	mapstructure.Decode(decoded.(jwt.MapClaims), &user)
// 	json.NewEncoder(w).Encode(user)
// }
