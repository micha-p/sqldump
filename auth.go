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

type Env struct {
	User string
	Pass string
	Host string
	Port string
	Dbms string
}

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func getCredentials(r *http.Request) (string, string, string, string, string, error) {

	var dbms, host, port, user, pass string

	cookie, err := r.Cookie("Datasource")
	if err == nil {
		cookieValue := make(map[string]string)
		err = cookieHandler.Decode("Datasource", cookie.Value, &cookieValue)
		if err == nil {
			dbms = cookieValue["dbms"]
			host = cookieValue["host"]
			port = cookieValue["port"]
			user = cookieValue["user"]
			pass = cookieValue["pass"]
		} else { // cookieerror
			log.Println("[Cookie error] ", user+"@"+host+":"+port+"("+dbms+")")
			return "", "", "", "", "", err
		}
	}
	return dbms, host, port, user, pass, err
}

func setCredentials(w http.ResponseWriter, r *http.Request, dbms string, host string, port string, user string, pass string) {
	value := map[string]string{
		"dbms": dbms,
		"host": host,
		"port": port,
		"user": user,
		"pass": pass,
	}
	if encoded, err := cookieHandler.Encode("Datasource", value); err == nil {
		c := &http.Cookie{
			Name:  "Datasource",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, c)
		if DEBUGFLAG {
			log.Println("[Cookie] " + user + "@" + host + ":" + port + "(" + dbms + ")")
		}
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
	db := request.FormValue("db")
	host := request.FormValue("host")
	port := request.FormValue("port")
	dbms := request.FormValue("dbms")
	if user != "" && pass != "" {
		log.Println("[LOGIN]", dbms, user, host, port, db)
		setCredentials(w, request, dbms, host, port, user, pass)
	}
	http.Redirect(w, request, request.URL.Host+"/?db="+db, 302)
}

func logoutHandler(w http.ResponseWriter, request *http.Request) {
	clearCredentials(w)
	http.Redirect(w, request, request.URL.Host, 302)
}
