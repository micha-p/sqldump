package main

import (
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"strconv"
)

var database = "information_schema"
var EXPERTFLAG bool
var INFOFLAG bool
var DEBUGFLAG bool
var READONLY bool
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
		log.Println("[GET]", CSS_FILE)
	} else {
		http.StatusText(404)
	}
}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, loginPage)
}

func readRequest(r *http.Request) (string, string, string, string, string, string, string) {
	q := r.URL.Query()
	db := sqlProtectIdentifier(q.Get("db"))
	t := sqlProtectIdentifier(q.Get("t"))
	o := sqlProtectIdentifier(q.Get("o"))
	d := sqlProtectIdentifier(q.Get("d"))
	k := sqlProtectIdentifier(q.Get("k"))
	n := sqlProtectString(q.Get("n"))
	v := sqlProtectString(q.Get("v"))
	return db, t, o, d, n, k, v
}

func workload(w http.ResponseWriter, r *http.Request, cred Access) {

	db, t, o, d, n, k, v := readRequest(r)

	q := r.URL.Query()
	action := q.Get("action")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if action == "SUBSET" && db != "" && t != "" {
		actionSubset(w, r, cred, db, t)
	} else if action == "QUERY" && db != "" && t != "" {
		actionQuery(w, r, cred, db, t)
	} else if action == "INFO" && db != "" && t != "" {
		actionInfo(w, r, cred, db, t)
	} else if action == "ADD" && !READONLY && db != "" && t != "" {
		actionAdd(w, r, cred, db, t)
	} else if action == "INSERT" && !READONLY && db != "" && t != "" {
		actionInsert(w, r, cred, db, t)
	} else if action == "REMOVE" && !READONLY && db != "" && t != "" && k != "" && v != "" {
		actionRemove(w, r, cred, db, t, k, v)
	} else if action == "EDIT" && !READONLY && db != "" && t != "" && k != "" && v != "" {
		actionEdit(w, r, cred, db, t, k, v)
	} else if action == "EDITEXEC" && !READONLY && db != "" && t != "" && k != "" && v != "" {
		actionEditExec(w, r, cred, db, t, k, v)
	} else if action == "DELETEFORM" && !READONLY && db != "" && t != "" { // Subset and Delete 1
		actionDeleteForm(w, r, cred, db, t)
	} else if action == "DELETEEXEC" && !READONLY && db != "" && t != "" { // Subset and Delete 2
		actionDeleteExec(w, r, cred, db, t)
	} else if action == "DELETE" && !READONLY && db != "" && t != "" { // Delete a selected subset
		actionDeleteSubset(w, r, cred, db, t)
	} else if action == "UPDATE" && !READONLY && db != "" && t != "" { // Update a selected subset
		actionUpdateSubset(w, r, cred, db, t)
	} else if action == "UPDATEEXEC" && !READONLY && db != "" && t != "" {
		actionUpdateExec(w, r, cred, db, t)
	} else if action == "GOTO" && db != "" && t != "" && n != "" {
		dumpIt(w, r, cred, db, t, o, d, n, k, v)
	} else if action == "BACK" {
		dumpIt(w, r, cred, db, "", "", "", "", "", "")
	} else if action != "" {
		shipMessage(w, cred, db, "Unknown action: "+action)
	} else {
		dumpIt(w, r, cred, db, t, o, d, n, k, v)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	log.Println("[GET]", r.URL)
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
	var READONLYFLAG = flag.Bool("r", false, "read-only access")

	flag.Parse()

	READONLY = *READONLYFLAG
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
