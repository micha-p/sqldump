package main

import (
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var database = "information_schema"
var EXPERTFLAG bool
var INFOFLAG bool
var DEBUGFLAG bool
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
		fmt.Println("get css: " + CSS_FILE)
	} else {
		http.StatusText(404)
	}
}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, loginPage)
}

func sqlprotect(s string) string {
	if s != "" && strings.ContainsAny(s, "\"\\;") {
		r := strings.Replace(strings.Replace(strings.Replace(s, "\\", "", -1), ";", "", -1), "\"", "", -1)
		log.Println("SQL INJECTION! :" + s + "->" + r)
		return r
	} else {
		return s
	}
}

func readRequest(request *http.Request) (string, string, string, string, string, string, string) {
	q := request.URL.Query()
	db := sqlprotect(q.Get("db"))
	t := sqlprotect(q.Get("t"))
	o := sqlprotect(q.Get("o"))
	d := sqlprotect(q.Get("d"))
	n := sqlprotect(q.Get("n"))
	k := sqlprotect(q.Get("k"))
	v := sqlprotect(q.Get("v"))
	return db, t, o, d, n, k, v
}

func workload(w http.ResponseWriter, r *http.Request, cred Access) {

	db, t, o, d, n, k, v := readRequest(r)
	q := r.URL.Query()
	action := q.Get("action")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if action == "subset" && db != "" && t != "" {
		actionSubset(w, r, cred, db, t)
	} else if action == "query" && db != "" && t != "" {
		actionQuery(w, r, cred)
	} else if action == "add" && db != "" && t != "" {
		actionAdd(w, r, cred, db, t)
	} else if action == "insert" && db != "" && t != "" {
		actionInsert(w, r, cred)
	} else if action == "info" && db != "" && t != "" {
		q.Del("action")
		actionInfo(w, r, cred, db, t, "?"+q.Encode())
	} else if action == "goto" && db != "" && t != "" && n != "" {
		dumpIt(w, cred, db, t, o, d, n, k, v)
	} else {
		dumpIt(w, cred, db, t, o, d, n, k, v)
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
	var DEBUG = flag.Bool("d", false, "dynamically load html templates and css (DEBUG)")

	flag.Parse()

	INFOFLAG = *INFO
	DEBUGFLAG = *DEBUG
	EXPERTFLAG = *EXPERT
	CSS_FILE = *CSS
	initTemplate()

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
		if CSS_FILE != "" {
			fmt.Println("using style in " + CSS_FILE)
		}
		if DEBUGFLAG {
			fmt.Println("dynamically loading html templates and css (DEBUG)")
		}
		http.ListenAndServeTLS(portstring, "cert.pem", "key.pem", nil)
	} else {
		fmt.Println("Listening at http://" + *HOST + portstring)
		if CSS_FILE != "" {
			fmt.Println("using style in " + CSS_FILE)
		}
		if DEBUGFLAG {
			fmt.Println("dynamically loading html templates and css (DEBUG)")
		}
		http.ListenAndServe(portstring, nil)
	}
}
