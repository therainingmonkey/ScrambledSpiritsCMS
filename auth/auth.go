package auth

import (
	"net/http"
	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64), // Hash key
	securecookie.GenerateRandomKey(32), // Block key
)

// Takes a string and retruns its hash as type []byte
func CreateHash(pass string) (passwordHash string) {
	if passwordHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost); err == nil {
		return string(passwordHash)
	} else {
		return ""
	}

}

// Returns true if the password matches the stored passwordHash, else return false
func CheckCredentials(passwordHash, pass string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(pass)); err == nil {
		return true
	} else {
		return false
	}
}

// Returns username if they have an authentication cookie, else returns and empty string
func GetUserName(req *http.Request) (userName string) {
	if cookie, err := req.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		} else {
			userName = ""
		}
	}
	return userName
}

// Sets an authentication cookie
func SetSession(userName string, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

// Clears the authentication cookie
func ClearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}
