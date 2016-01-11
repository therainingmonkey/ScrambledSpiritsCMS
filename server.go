/*
TODO: Interface for making / editing posts
TODO: Write content (band blurbs etc.)
TODO: Make pages prettier (CSS, Images)
*/
package main

import (
	"io"
	"html/template"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	// Local Packages
	"./models"
)

const STATIC_URL string = "static/"
const STATIC_ROOT string = "static/"

var db gorm.DB

type Context struct {
	Title string
	StaticPath string
	Posts []models.Post
}

func checkErr(err error, message string){
	if err != nil{
		panic(message) // TODO Change to log.Fatal(message)
	}
}

func initDB() (err error) {
	db, err = gorm.Open("sqlite3", "test.db") // TODO change db name
//	db.LogMode(true) // DEBUG
	db.CreateTable(&models.Post{})
	db.AutoMigrate()
	return err
}

// TODO change to func render, to be called from other page handlers.
func render(w http.ResponseWriter, context Context) {
	context.StaticPath = STATIC_URL
	t, err := template.ParseFiles("templates/main.html")
	checkErr(err, "Could not load template")
	err = t.Execute(w, context)
	checkErr(err, "Could not execute template")
}

//TODO separate context from server logic.
func home(w http.ResponseWriter, req * http.Request) {
	context := Context{Title: "Scrambled Spirits Collective"}
	db.Order("created_at desc").Find(&context.Posts)
	render(w, context)
}

func staticHandler(w http.ResponseWriter, req *http.Request) {
	file_name := req.URL.Path[len(STATIC_URL):] // Get filename from URL
	if len(file_name) != 0 {
		f, err := http.Dir(STATIC_ROOT).Open(file_name)
		if err == nil {
			content := io.ReadSeeker(f)
			http.ServeContent(w, req, file_name, time.Now(), content)
			return
		}
	}
	http.NotFound(w, req)
}

func main(){
	err := initDB()
	checkErr(err, "Could not open database")

	http.HandleFunc("/", home)
	http.HandleFunc("/static/", staticHandler)
	http.ListenAndServe(":8080", nil)
}
