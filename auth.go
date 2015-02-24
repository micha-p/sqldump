/* no session management needed
 * Credentials are stored at user side using secure cookies
 *
 * credits:
 * http://www.mschoebel.info/2014/03/09/snippet-golang-webapp-login-logout.html
 */

package main

import (
	"github.com/gorilla/securecookie"
	"net/http"
	"fmt"
)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func getCredentials(r *http.Request) (user string, pass string, host string, port string) {
	
	cookie, err := r.Cookie("Datasource")
	if err == nil {
		cookieValue := make(map[string]string)
		err = cookieHandler.Decode("Datasource", cookie.Value, &cookieValue)
		if err == nil {
			user = cookieValue["user"]
			pass = cookieValue["pass"]
			host = cookieValue["host"]
			port = cookieValue["port"]
			fmt.Println("Credentials taken for " + host + ":" + port)
		} else { // cookieerror
			fmt.Println("Cookie error " + host + ":" + port)
		}
	} 	
	return user, pass, host, port
}

func checkCredentials(r *http.Request) error {
	
	cookie, err := r.Cookie("Datasource")
	if err == nil {
		cookieValue := make(map[string]string)
		return cookieHandler.Decode("Datasource", cookie.Value, &cookieValue)			
	} else {
		return err
	}
}

func setCredentials(w http.ResponseWriter, r *http.Request, user string, pass string, host string, port string) *http.Request{
	value := map[string]string{
		"user":   user,
		"pass":   pass,
		"host":   host,
		"port":   port,
	}
	if encoded, err := cookieHandler.Encode("Datasource", value); err == nil {
		c := &http.Cookie{
			Name:  "Datasource",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, c)
		fmt.Println("Credentials set for " + user + "@" + host + ":" + port)
		r.AddCookie(c)
	}
	return r
}

func clearCredentials(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "Datasource",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

func loginHandler(w http.ResponseWriter, request *http.Request) {
	user := request.FormValue("user")
	pass := request.FormValue("pass")
	host := request.FormValue("host")
	port := request.FormValue("port")
	if user != "" && pass != "" {
		setCredentials(w, request, user, pass, host, port)
	}
	http.Redirect(w, request, request.URL.Host, 302)
}

func logoutHandler(w http.ResponseWriter, request *http.Request) {
	clearCredentials(w)
	http.Redirect(w, request, request.URL.Host, 302)
}
