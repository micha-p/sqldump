package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
)

var database = "information_schema"

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.StatusText(404)
}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, loginPage)
}

func helpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, helpPage)

}

func workload(w http.ResponseWriter, r *http.Request) {

	v := r.URL.Query()
	action := v.Get("action")
	db := v.Get("db")
	t := v.Get("t")

	if action == "select" && db != "" && t != "" {
		actionSelect(w, r, db, t)
	} else if action == "insert" && db != "" && t != "" {
		actionInsert(w, r, db, t)
	} else {
		dumpIt(w, r)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	pass := ""
	user, _, host, port := getCredentials(r)

	if user != "" {
		workload(w, r)
	} else {
		v := r.URL.Query()
		user = v.Get("user")
		pass = v.Get("pass")
		host = v.Get("host")
		port = v.Get("port")

		if user != "" && pass != "" {
			if host == "" {
				host = "localhost"
			}
			if port == "" {
				port = "3306"
			}
			setCredentials(w, user, pass, host, port)
			http.Redirect(w, r, r.URL.Host, 302)
		} else {
			loginPageHandler(w, r)
		}
	}
}

func main() {

	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/help", helpHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/insert", insertHandler)
	http.HandleFunc("/", indexHandler)

	fmt.Println("Listening at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
