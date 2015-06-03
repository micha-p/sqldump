/* no session management needed
 * Credentials are stored at user side using secure cookies
 *
 * credits:
 * http://www.mschoebel.info/2014/03/09/snippet-golang-webapp-login-logout.html
 */

package main

import (
	"github.com/gorilla/securecookie"
	"log"
	"net/http"
)

type Access struct {
	User string
	Pass string
	Host string
	Port string
	Dbms string
}

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func getCredentials(r *http.Request) (Access, error) {

	var user, pass, host, port, dbms string

	cookie, err := r.Cookie("Datasource")
	if err == nil {
		cookieValue := make(map[string]string)
		err = cookieHandler.Decode("Datasource", cookie.Value, &cookieValue)
		if err == nil {
			user = cookieValue["user"]
			pass = cookieValue["pass"]
			host = cookieValue["host"]
			port = cookieValue["port"]
			dbms = cookieValue["dbms"]
		} else { // cookieerror
			log.Println("[Cookie error] ", host+":"+port+"("+dbms+")")
			return Access{user, pass, host, port, dbms}, err
		}
	}
	return Access{user, pass, host, port, dbms}, err
}

func setCredentials(w http.ResponseWriter, r *http.Request, cred Access) {
	value := map[string]string{
		"user": cred.User,
		"pass": cred.Pass,
		"host": cred.Host,
		"port": cred.Port,
		"dbms": cred.Dbms,
	}
	if encoded, err := cookieHandler.Encode("Datasource", value); err == nil {
		c := &http.Cookie{
			Name:  "Datasource",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, c)
		log.Println("[Cookie] " + cred.User + "@" + cred.Host + ":" + cred.Port + "(" + cred.Dbms + ")")
	}
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
	dbms := request.FormValue("dbms")
	if user != "" && pass != "" {
		setCredentials(w, request, Access{user, pass, host, port, dbms})
	}
	http.Redirect(w, request, request.URL.Host, 302)
}

func logoutHandler(w http.ResponseWriter, request *http.Request) {
	clearCredentials(w)
	http.Redirect(w, request, request.URL.Host, 302)
}
