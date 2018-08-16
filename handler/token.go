package handler

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
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/sikajs/my-go-api/db"
	"github.com/sikajs/my-go-api/model"
)

const (
	secretKey = "Welcome to JS's playground"
)

//GetToken let client get a new token to use
func GetToken(w http.ResponseWriter, r *http.Request) {
	var v map[string]interface{}
	var user model.User
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

//ValidateMiddleware adds token checking
func ValidateMiddleware(next http.HandlerFunc) http.HandlerFunc {
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
