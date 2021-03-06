package model

import (
	"github.com/jinzhu/gorm"
)

// Post data structure
type Post struct {
	gorm.Model
	Title    string
	Content  string
	Author   User `gorm:"foreignkey:AuthorID;association_foreignkey:ID"`
	AuthorID uint
}

// User data structure
type User struct {
	gorm.Model
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}
