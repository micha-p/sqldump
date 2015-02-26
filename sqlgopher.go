package main

import (
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"strconv"
)

var database = "information_schema"
var EXPERTFLAG bool
var INFOFLAG bool
var CSS_FILE string

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	if troubleF("favicon.ico") == nil {
		http.ServeFile(w, r, "favicon.ico")
	} else {
		http.StatusText(404)
	}
}

func cssHandler(w http.ResponseWriter, r *http.Request) {
	if troubleF(CSS_FILE) == nil {
		http.ServeFile(w, r, CSS_FILE)
	} else {
		http.StatusText(404)
	}
}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, loginPage)
}

func workload(w http.ResponseWriter, r *http.Request, cred Access) {

	q := r.URL.Query()
	action := q.Get("action")
	db := q.Get("db")
	t := q.Get("t")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if action == "subset" && db != "" && t != "" {
		actionSubset(w, r, cred, db, t)
	} else if action == "query" && db != "" && t != "" {
		actionQuery(w, r, cred)
	} else if action == "add" && db != "" && t != "" {
		actionAdd(w, r, cred, db, t)
	} else if action == "insert" && db != "" && t != "" {
		actionInsert(w, r, cred)
	} else if action == "show" && db != "" && t != "" {
		q.Del("action")
		actionShow(w, r, cred, db, t, "?"+q.Encode())
	} else {
		dumpIt(w, r, cred)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	fmt.Println(r.URL)
	user := q.Get("user")
	pass := q.Get("pass")
	host := q.Get("host")
	port := q.Get("port")
	dbms := q.Get("dbms")

	if user != "" && pass != "" {
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "3306"
		}
		if dbms == "" {
			dbms = "mysql"
		}
		cred := Access{user, pass, host, port, dbms}
		setCredentials(w, r, cred)
		workload(w, r, cred)

	} else {
		if cred, err := getCredentials(r); err == nil {
			workload(w, r, cred)
		} else {
			loginPageHandler(w, r)
		}
	}
}

func main() {

	var SECURE = flag.Bool("s", false, "https Connection TLS")
	var HOST = flag.String("h", "localhost", "server name")
	var PORT = flag.Int("p", 8080, "server port")
	var INFO = flag.Bool("i", false, "include INFORMATION_SCHEMA in overview")
	var EXPERT = flag.Bool("x", false, "expert mode to access privileges, routines, triggers, views (TODO)")
	var CSS = flag.String("c", "", "supply customized style in CSS file")

	flag.Parse()

	INFOFLAG = *INFO
	EXPERTFLAG = *EXPERT
	CSS_FILE = *CSS

	portstring := ":" + strconv.Itoa(*PORT)

	if CSS_FILE != "" && troubleF(CSS_FILE) == nil {
		http.HandleFunc("/"+CSS_FILE, cssHandler)
	}
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/", indexHandler)

	if *SECURE {
		if troubleF("cert.pem") == nil && troubleF("key.pem") == nil {
			fmt.Println("cert.pem and key.pem found")
		} else {
			fmt.Println("Generating cert.pem and key.pem ...")
			generate_cert(*HOST, 2048, false)
		}
		fmt.Println("Listening at https://" + *HOST + portstring)
		http.ListenAndServeTLS(portstring, "cert.pem", "key.pem", nil)
	} else {
		fmt.Println("Listening at http://" + *HOST + portstring)
		http.ListenAndServe(portstring, nil)
	}
}
