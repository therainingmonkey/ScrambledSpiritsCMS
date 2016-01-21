package models

import (
	"time"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

// Base model, inherited by others
type Model struct {
	ID			uint 		`gorm:"primary_key"`
	CreatedAt	time.Time
	UpdatedAt	time.Time
	DeletedAt	time.Time
}

// Posts table
type Post struct {
	gorm.Model	// inherits Model struct
	Title	string
	Body	string
	Author	string
	Tags	[]Tag
}

// Tag table, one table for each tag
type Tag struct {
	gorm.Model
	PostID	uint
}

type User struct {
	gorm.Model
	Username	string
	PasswordHash	string
}
