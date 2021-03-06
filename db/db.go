package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sikajs/my-go-api/model"
)

type config struct {
	Host     string
	Port     int
	Username string
	Password string
	Name     string
	SSLMode  string
}

func getDbConfig() *config {
	return &config{
		Host:     "localhost",
		Port:     5432,
		Username: "pqgotest",
		Password: "password",
		Name:     "pqgotest",
		SSLMode:  "disable",
	}
}

// Connect with db config
func Connect() *sql.DB {
	// config := getDbConfig()
	// connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", config.Host, config.Port, config.Username, config.Password, config.Name, config.SSLMode)
	// db, err := sql.Open("postgres", connStr)
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(`Database connected.`)
	return db
}

// GormConn connect db with Gorm
func GormConn() *gorm.DB {
	config := getDbConfig()
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", config.Host, config.Port, config.Username, config.Password, config.Name, config.SSLMode)
	db, err := gorm.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&model.Post{}, &model.User{})

	fmt.Println(`Database connected.`)
	return db
}
