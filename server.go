/*
TODO: Write content (band blurbs etc.) (Twitter feed? FB like link?)
TODO: Make pages prettier (CSS, Images)
TODO: button to upload images & embed in posts.
TODO: Paginate posts
TODO: Make the login template use <keygen> fields (http basic authentication?)
TODO: Add RSS feed (gorilla/feeds?)
*/
package main

import (
	"io"
	"io/ioutil"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	// Local Packages
	"./auth"
	"./models"
)

var (
	db gorm.DB
	ArtistContextMap = map[string]*Context {
		"thecalamity": &Context {Title: "The Calamity"},
		"aweatherman": &Context {Title: "A Weatherman"},
		"bingolittle": &Context {Title: "Bingo Little"},
		"figurinesofthewretched": &Context {Title: "Figurines of the Wretched"},
	}
)

type Context struct {
	Title string
	Username string
	StaticText template.HTML
	Posts []models.Post
}


func checkErr(err error, message string) {
	if err != nil{
		panic(message + err.Error()) // TODO Change to log.Fatal(message)
	}
}

func buildArtistMap() {	// TODO: Make this work
	for artistName, _ := range ArtistContextMap {
		text, err := ioutil.ReadFile("static/sitetext/" + artistName)
		checkErr(err, "Could not load static text")
		ArtistContextMap[artistName].StaticText = template.HTML(string(text))
	}
}

func initDB() (err error) {
	// TODO Check whether tables exist before creating them
	db, err = gorm.Open("sqlite3", "test.db") // TODO change db name
	checkErr(err, "Could not open databse")
//	db.LogMode(true) // DEBUG
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

func checkTagCheckbox(tagstr string, postID uint, req *http.Request) error {
	if len(req.FormValue(tagstr)) > 0 {
		dbtag := models.Tag{PostID: postID}
		err := db.Table("Tag" + tagstr).Create(&dbtag).Error
		return err
	} else {
		return nil
	}
}

func homeHandler(w http.ResponseWriter, req * http.Request) {
	username := auth.GetUserName(req)
	text, err := ioutil.ReadFile("static/sitetext/home")
	checkErr(err, "Could not load static text")
	context := Context{
		Title: "Scrambled Spirits Collective",
		Username: username,
		StaticText: template.HTML(string(text)),
					  } // TODO Read from file, add content.
	db.Order("created_at desc").Find(&context.Posts) // Fills context.Posts with posts from the database.
	err = render(w, context, "templates/main.html")
	checkErr(err, "Problem generating homepage")
}

func artistHandler(w http.ResponseWriter, req *http.Request) {
	artistName := req.URL.Path[len("/artist/"):] // Get artist name from URL
	if len(artistName) != 0 {
		context, ok := ArtistContextMap[artistName] // OK is true if the key exists, false otherwise
		if ok {
			tags := []models.Tag{}
			db.Table("Tag" + artistName).Order("created_at desc").Find(&tags) // Fills tags
			for _, tag := range(tags) { // Find the tag's associated post and add it to context.Posts
				post := models.Post{}
				err := db.Model(&tag).Related(&post).Error
				checkErr(err, "Problem finding tagged posts.")
				context.Posts = append(context.Posts, post)
			} // TODO: Can this be structured better? Pagination? More within the idiom of GORM?
			err := render(w, *context, "templates/main.html")
			checkErr(err, "Problem generating artist page")
			return
		} else {
			http.NotFound(w, req)
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

// TODO Restructure - subfunctions (error handling)
func editHandler(w http.ResponseWriter, req *http.Request) {
	context := Context{Username: auth.GetUserName(req)}
	context.Posts = append(context.Posts, models.Post{})
	if context.Username != "" { // If they have an authentication cookie
		post := &context.Posts[0]
		if len(req.URL.Path[len("/edit/"):]) > 0 { // If a post ID is referred to in the URL
			urlID, err := strconv.Atoi(req.URL.Path[len("/edit/"):]) //Get the target post ID from URL
			checkErr(err, "Could not get postID from URL")
			post.ID = uint(urlID)
			db.FirstOrCreate(post) // If db contains post.ID, copy the entry to post; else create an entry
		}
		switch req.Method {
			case "GET":
				err := render(w, context, "templates/edit.html")
				checkErr(err, "Could not render Edit template")
			case "POST":
			// TODO look into "bind" library for getting form values & converting to/from time.Time
				if len(req.FormValue("deleteButton")) > 0 { // Check whether they pressed the "delete" button
					db.Table("posts").Delete(post)
					for artistName, _ := range ArtistContextMap {
						db.Table("Tag" + artistName).Where("post_id = ?", post.ID).Delete(models.Tag{})
					}
				} else {
					post.Title = req.FormValue("title")
					post.Author = req.FormValue("author")
					post.Body = req.FormValue("body")
					db.Save(&post)
					for key := range ArtistContextMap {
						err := checkTagCheckbox(key, post.ID, req)
						checkErr(err, "Problem parsing 'tag' checkboxes.")
					}
				}
			http.Redirect(w, req, "/", 302)
		}
	} else { // If no auth cookie
		http.NotFound(w, req)
	}
}

func logoutHandler (w http.ResponseWriter, req *http.Request) {
	auth.ClearSession(w)
	http.Redirect(w, req, "/", 302)
}

/* TODO Add users etc.
func adminHandler(w http.ResponseWriter, req *http.Request) {

}*/

func staticHandler(w http.ResponseWriter, req *http.Request) {
	file_name := req.URL.Path[len("static/"):] // Get filename from URL
	if len(file_name) != 0 {
		f, err := http.Dir("static/").Open(file_name)
		if err == nil {
			content := io.ReadSeeker(f)
			http.ServeContent(w, req, file_name, time.Now(), content)
			return
		}
	}
	http.NotFound(w, req)
}

func main(){
	buildArtistMap()
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
