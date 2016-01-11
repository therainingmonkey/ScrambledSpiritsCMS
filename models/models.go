package models

import (
	"time"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

type Model struct {
	ID			uint 		`gorm:"primary_key"`
	CreatedAt	time.Time
	UpdatedAt	time.Time
	DeletedAt	time.Time
}

type Post struct {
	gorm.Model	// inherits Model struct
	Title	string
	Body	string
	Author	string
}
