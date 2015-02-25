package main

import (
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"fmt"
	"flag"
	"strconv"
)

var database = "information_schema"
var EXPERTFLAG bool
var INFOFLAG bool

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

	q := r.URL.Query()
	fmt.Println(r.URL)
	user := q.Get("user")
	pass := q.Get("pass")
	host := q.Get("host")
	port := q.Get("port")

	if user != "" && pass != "" {
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "3306"
		}
		newrequest := setCredentials(w, r, user, pass, host, port)
		workload(w, newrequest)
		
	} else {
		if checkCredentials(r) == nil {
			workload(w, r)
		} else {
			loginPageHandler(w, r)
		}
	}
}


func main() {

	var SECURE = flag.Bool ("s", false, "https Connection TLS")
	var HOST = flag.String ("h", "localhost", "server name")
	var PORT = flag.Int ("p", 8080, "server port")
	var INFO = flag.Bool ("i", false, "include INFORMATION_SCHEMA in overview")
	var EXPERT = flag.Bool ("x", false, "expert mode to access privileges, routines, triggers, views (TODO)")
	var CSS	   = flag.String("c","", "supply customized style in CSS file") // TODO

	flag.Parse()

	INFOFLAG = *INFO
	EXPERTFLAG = *EXPERT
	CSS_FILE = *CSS

	portstring := ":" + strconv.Itoa(*PORT)
	
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/", indexHandler)

	if *SECURE {
		if troubleF("cert.pem")==nil && troubleF("key.pem")==nil  {
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
