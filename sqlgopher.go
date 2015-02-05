package main

import (
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"fmt"
)

var database = "information_schema"

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.StatusText(404)
}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, loginPage)
}


func workload(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	action := q.Get("action")
	db := q.Get("db")
	t := q.Get("t")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if action == "subset" && db != "" && t != "" {
		actionSubset(w, r, db, t)
	} else if action == "query" && db != "" && t != "" {
		actionQuery(w, r)
	} else if action == "add" && db != "" && t != "" {
		actionAdd(w, r, db, t)
	} else if action == "Insert" && db != "" && t != "" {
		actionInsert(w, r)
	} else if action == "show" && db != "" && t != "" {
		q.Del("action")
		actionShow(w, r, db, t, "?"+q.Encode())
	} else {
		dumpIt(w, r)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	if checkCredentials(r) == nil {
		workload(w, r)
	} else {
		q := r.URL.Query()
		user := q.Get("user")
		pass := q.Get("pass")
		host := q.Get("host")
		port := q.Get("port")
		q.Del("user")
		q.Del("pass")
		q.Del("host")
		q.Del("port")

		if user != "" && pass != "" {
			if host == "" {
				host = "localhost"
			}
			if port == "" {
				port = "3306"
			}
			setCredentials(w, r, user, pass, host, port)
			workload(w, r)
		} else {
			loginPageHandler(w, r)
		}
	}
}


func main() {

	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/", indexHandler)

	fmt.Println("Listening at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
