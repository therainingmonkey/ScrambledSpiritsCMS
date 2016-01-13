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

var ArtistContextMap = map[string]ArtistContext {
	"thecalamity": ArtistContext {
		About: "The Calamity are a band."
	}
}

type Context struct {
	Title string
	StaticPath string
	Posts []models.Post
}

type ArtistContext struct {
	Context
	About string
}


func checkErr(err error, message string){
	if err != nil{
		panic(message) // TODO Change to log.Fatal(message)
	}
}

func initDB() (err error) {
	db, err = gorm.Open("sqlite3", "test.db") // TODO change db name
	db.CreateTable(&models.Post{}) // Create a table for posts
	// Create a table for each tag linking to posts with that tag
	for artistName, _ := range ArtistContextMap {
		db.Table("Tag" + artistName).CreateTable(&models.Tag) // Append "tag" to artistName to avoid future collisions
	}
	db.AutoMigrate() //DEBUG ?
	return err
}

func render(w http.ResponseWriter, context Context) {
	context.StaticPath = STATIC_URL
	t, err := template.ParseFiles("templates/main.html")
	checkErr(err, "Could not load template")
	err = t.Execute(w, context)
	checkErr(err, "Could not execute template")
}

//TODO separate context from server logic?
func homeHandler(w http.ResponseWriter, req * http.Request) {
	context := Context{Title: "Scrambled Spirits Collective"}
	db.Order("created_at desc").Find(&context.Posts) // Fills context.Posts with posts from the database.
	render(w, context)
}

func artistHandler(w http.ResponseWriter, req *http.Request) {
	artistName := req.URL.Path["len(artist/"):] // Get artist name from URL
	if len(artistName) != 0 {
		context, ok := ArtistContextMap[artistName] // OK is true if the key exists, false otherwise
		if ok {
			// Lookup tag from artistcontextmap
			db.Limit(10).Order("created_at desc").Table("Tag" + artistName).Association("Post").Find(&context.Posts)
			// TODO ++ Update database schema & queries for tag tables.
			render(w, context)
		}
	}
	http.NotFound(w, req)
}

/* TODO
func adminHandler(w http.ResponseWriter, req *http.Request) {

}*/

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

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/artist/", artistHandler)
	http.HandleFunc("/static/", staticHandler)
	http.HandleFunc("/admin/", adminHandler)
	http.ListenAndServe(":8080", nil)
}
