/*
TODO: Authentication for editors/admin
TODO: Interface for making / editing posts
TODO: Write content (band blurbs etc.) (Twitter feed? FB like link?)
TODO: Make pages prettier (CSS, Images)
TODO: Add RSS feed (gorilla/feeds)
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
	"./auth"
	"./models"
)

const (
	STATIC_ROOT string = "static/"
)

var (
	db gorm.DB
	ArtistContextMap = map[string]Context {"thecalamity": Context {StaticText: "The Clamity are a acoostik band."}} // TODO Content
)

type Context struct {
	Title string
	Username string
	StaticText string
	Posts []models.Post
}


func checkErr(err error, message string){
	if err != nil{
		panic(message) // TODO Change to log.Fatal(message)
	}
}

func initDB() (err error) {
	// TODO Check whether tables exist before creating them
	db, err = gorm.Open("sqlite3", "test.db") // TODO change db name
	checkErr(err, "Could not open databse")
	db.CreateTable(&models.Post{}) // Create a table for posts
	db.CreateTable(&models.User{}) // Create a table for users
	// Create a table for each tag linking to posts with that tag
	for artistName, _ := range ArtistContextMap {
		db.Table("Tag" + artistName).CreateTable(&models.Tag{}) // Append "tag" to artistName to avoid future collisions
	}
	db.AutoMigrate()
	return err
}

func render(w http.ResponseWriter, context Context, tmpl string) error {
	t, err := template.ParseFiles(tmpl, "templates/head.html", "templates/foot.html")
	if err != nil {
		return err
	}
	err = t.ExecuteTemplate(w, "content", context)
	return err
}

//TODO separate context from server logic? error handling,
func homeHandler(w http.ResponseWriter, req * http.Request) {
	username := auth.GetUserName(req)
	context := Context{
		Title: "Scrambled Spirits Collective",
		Username: username,
		StaticText: "SSC 4 lyf 2k6teen!",
					  } // TODO Content
	db.Order("created_at desc").Find(&context.Posts) // Fills context.Posts with posts from the database.
	err := render(w, context, "templates/main.html")
	checkErr(err, "Problem generating homepage")
}

func artistHandler(w http.ResponseWriter, req *http.Request) {
	artistName := req.URL.Path[len("/artist/"):] // Get artist name from URL
	if len(artistName) != 0 {
		context, ok := ArtistContextMap[artistName] // OK is true if the key exists, false otherwise
		if ok {
			// Lookup tag from artistcontextmap
			db.Limit(10).Order("created_at desc").Table("Tag" + artistName).Association("Post").Find(&context.Posts) // FIXME
			err := render(w, context, "templates/main.html")
			checkErr(err, "Problem generating artist page")
		}
	}
	http.NotFound(w, req)
}

// TODO DRY, Error handling, ESCAPE INPUT, standardise redirects
func loginHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
		case "POST":
			username := req.FormValue("username")
			pass := req.FormValue("password")
			if username != "" && pass != "" {
				user := new(models.User)
				if err := db.Where("username = ?", username).First(&user).Error; err == gorm.RecordNotFound {
					http.NotFound(w, req)
					break
				} else if auth.CheckCredentials(user.PasswordHash, pass) {
					auth.SetSession(username, w)
				}
				http.Redirect(w, req, "/", 302)
			}
		default:
			context := Context{Title: "Login"}
			err := render(w, context, "templates/loginform.html")
			checkErr(err, "Problem generating login page")
	}
}

func editHandler(w http.ResponseWriter, req *http.Request) {
	username := auth.GetUserName(req)
	if username != "" { // If they have an authentication cookie
		switch req.Method {
			case "GET":
				context := Context{Username: username}
				postID := req.URL.Path[len("/edit/"):]
				if err := db.Where("ID = ?", postID).First(&context.Posts).Error;  err == gorm.RecordNotFound {
					http.NotFound(w, req)
					break
				}
				err := render(w, context, "templates/edit.html")
				checkErr(err, "Could not render Edit template")
			case "POST":
			// TODO ++ DB operations
		}
	} else {
		http.NotFound(w, req)
	}
}

func logoutHandler (w http.ResponseWriter, req *http.Request) {
	auth.ClearSession(w)
	http.Redirect(w, req, "/", 302)
}

/* TODO Add users etc.
func adminHandler(w http.ResponseWriter, req *http.Request) {

}
*/
func staticHandler(w http.ResponseWriter, req *http.Request) {
	file_name := req.URL.Path[len("static/"):] // Get filename from URL
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
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/logout/", logoutHandler)
//	http.HandleFunc("/admin/", adminHandler)
	http.ListenAndServe(":8080", nil)
}
