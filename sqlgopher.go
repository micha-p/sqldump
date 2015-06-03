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
		log.Println("[GET]", CSS_FILE)
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
		log.Println("[SQL INJECTION]", s+"->"+r)
		return r
	} else {
		return s
	}
}

func sqlprotectweak(s string) string {
	if s != "" && strings.ContainsAny(s, "\\;") {
		r := strings.Replace(strings.Replace(s, ";", "", -1), "\"", "", -1)
		log.Println("[SQL INJECTION]", s+"->"+r)
		return r
	} else {
		return s
	}
}

func readRequest(request *http.Request) (string, string, string, string, string, string, string, string) {
	q := request.URL.Query()
	db := sqlprotect(q.Get("db"))
	t := sqlprotect(q.Get("t"))
	o := sqlprotect(q.Get("o"))
	d := sqlprotect(q.Get("d"))
	n := sqlprotect(q.Get("n"))
	k := sqlprotect(q.Get("k"))
	v := sqlprotect(q.Get("v"))
	w := sqlprotectweak(q.Get("where"))
	return db, t, o, d, n, k, v, w
}

func workload(w http.ResponseWriter, r *http.Request, cred Access) {

	db, t, o, d, n, k, v, where := readRequest(r)
	q := r.URL.Query()
	action := q.Get("action")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if action == "SUBSET" && db != "" && t != "" {
		actionSubset(w, r, cred, db, t)
	} else if action == "QUERY" && db != "" && t != "" {
		actionQuery(w, r, cred, db, t)
	} else if action == "ADD" && db != "" && t != "" {
		actionAdd(w, r, cred, db, t)
	} else if action == "INSERT" && db != "" && t != "" {
		actionInsert(w, r, cred, db, t)
	} else if action == "INFO" && db != "" && t != "" {
		actionInfo(w, r, cred, db, t)
	} else if action == "REMOVE" && db != "" && t != "" && k != "" && v != "" {
		actionRemove(w, r, cred, db, t, k, v)
	} else if action == "EDIT" && db != "" && t != "" && k != "" && v != "" {
		actionEdit(w, r, cred, db, t, k, v)
	} else if action == "EDITEXEC" && db != "" && t != "" && k != "" && v != "" {
		actionEditExec(w, r, cred, db, t, k, v)
	} else if action == "DELETEWHERE" && db != "" && t != "" && where != "" {
		actionDeleteWhere(w, r, cred, db, t, where)
	} else if action == "DELETE" && db != "" && t != "" {
		actionDelete(w, r, cred, db, t)
	} else if action == "DELETEEXEC" && db != "" && t != "" {
		actionDeleteExec(w, r, cred, db, t)
	} else if action == "UPDATE" && db != "" && t != "" {
		shipMessage(w, cred, db, "Update records not implemented")
	} else if action == "GOTO" && db != "" && t != "" && n != "" {
		dumpIt(w, cred, db, t, o, d, n, k, v, where)
	} else if action == "BACK" {
		dumpIt(w, cred, db, "", "", "", "", "", "", "")
	} else if action != "" {
		shipMessage(w, cred, db, "Unknown action: "+action)
	} else {
		dumpIt(w, cred, db, t, o, d, n, k, v, where)
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
