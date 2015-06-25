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
	"fmt"
)


var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func getCredentialsFromRequest(r *http.Request) (string, string, string, string, string, string, error) {
	
	var err error
	q := r.URL.Query()
	log.Println("[GET]", r.URL)
	user := q.Get("user") // Login via bookmark
	pass := q.Get("pass")
	host := q.Get("host")
	port := q.Get("port")
	dbms := q.Get("dbms")
	base := q.Get("db")
	if dbms == "" {
		dbms = "mysql"
	}
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "3306"
	}
	if user == "" || pass == "" || base == "" {
		err= fmt.Errorf("Credentials incomplete for user '"+user+"'@'"+host+"' to database '"+base+"'")
	}
	return dbms, user, pass, host, port,base, err
}


func getCredentialsFromCookie(r *http.Request) (string, string, string, string, string, string, error) {

	var dbms, host, port, user, pass, base string

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
			base = cookieValue["base"]
			if DEBUGFLAG {
				log.Println("[HTTP]","Cookie found:", "(" + dbms + ") " + user + "@" + host + ":" + port + "/" + base )
			}
		} else { // cookieerror
			log.Println("[Cookie error] ", user+"@"+host+":"+port+"("+dbms+") "+base)
			return "", "", "", "", "", "", err
		}
	}
	return dbms, user, pass, host, port,base, err
}

func setCredentials(w http.ResponseWriter, r *http.Request, dbms string, user string, pass string, host string, port string, base string) {
	value := map[string]string{
		"dbms": dbms,
		"host": host,
		"port": port,
		"user": user,
		"pass": pass,
		"base": base,
	}
	if encoded, err := cookieHandler.Encode("Datasource", value); err == nil {
		c := &http.Cookie{
			Name:  "Datasource",
			Value: encoded,
			Path:  "/",    			// valid throughout the entire domain
		}
		http.SetCookie(w, c)
		if DEBUGFLAG {
			log.Println("[HTTP]","Cookie set:", "(" + dbms + ") " + user + "@" + host + ":" + port + "/" + base )
		}
	}
}

func clearCredentials(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "Datasource",
		Value:  "",
		Path:   "/",
		MaxAge: -1,					//	cookie’s expiration as an interval of seconds in the future
	}
	http.SetCookie(w, cookie)
	if DEBUGFLAG {
		log.Println("[HTTP]","Cookie withdrawn")
	}
}

func loginFormHandler(w http.ResponseWriter, request *http.Request) {
	user := request.FormValue("user")
	pass := request.FormValue("pass")
	base := request.FormValue("base")
	host := request.FormValue("host")
	port := request.FormValue("port")
	dbms := request.FormValue("dbms")
	if user != "" && pass != "" && base !="" {
		startWork(w, request, dbms, user, pass, host, port, base, true)
	}
}

func logoutHandler(w http.ResponseWriter, request *http.Request) {
	log.Println("[AUTH]","Logout")
	clearCredentials(w)
	http.Redirect(w, request, request.URL.Host, 302)
}
