package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"strings"
	"strconv"
)

// restrict GET to reserved files
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

func readRequest(r *http.Request) (string, string, string, string, string, string, string, string) {
	q := r.URL.Query()
	db := q.Get("db")
	t := q.Get("t")
	o := q.Get("o")
	d := q.Get("d")
	n := q.Get("n")
	g := q.Get("g")
	k := q.Get("k")
	v := q.Get("v")
	return db, t, o, d, n, g, k, v
}

func workload(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string) {

	db, t, o, d, n, g, k, v := readRequest(r)

	q := r.URL.Query()
	action := q.Get("action")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if action == "QUERY" && db != "" && t != "" {
		actionQUERY(w, r, conn, host, db, t, o, d)
	} else if action == "SELECT" && db != "" && t != "" {
		actionSELECT(w, r, conn, host, db, t, o, d)
	} else if action == "INFO" && db != "" && t != "" {
		actionINFO(w, r, conn, host, db, t)
	} else if action == "ADD" && !READONLY && db != "" && t != "" {
		actionADD(w, r, conn, host, db, t, o, d)
	} else if action == "INSERT" && !READONLY && db != "" && t != "" {
		actionINSERT(w, r, conn, host, db, t, o, d)
	} else if action == "QUERYDELETE" && !READONLY && db != "" && t != "" { // Create subset for DELETE
		actionQUERYDELETE(w, r, conn, host, db, t, o, d)
	} else if action == "DELETE" && !READONLY && db != "" && t != "" { // DELETE a selected subset
		actionDELETE(w, r, conn, host, db, t, o, d)
	} else if action == "UPDATE" && !READONLY && db != "" && t != "" { // UPDATE a selected subset
		actionUPDATE(w, r, conn, host, db, t, o, d)
	} else if action == "UPDATEFORM" && !READONLY && db != "" && t != "" { // ask for changed values
		actionUPDATEFORM(w, r, conn, host, db, t, o, d)
	} else if action == "EDITFORM" && !READONLY && db != "" && t != "" && k != "" && v != "" {
		actionEDITFORM(w, r, conn, host, db, t, k, v)
	} else if action == "UPDATEPRI" && !READONLY && db != "" && t != "" && k != "" && v != "" {
		actionUPDATEPRI(w, r, conn, host, db, t, k, v)
	} else if action == "DELETEPRI" && !READONLY && db != "" && t != "" && k != "" && v != "" {
		actionDELETEPRI(w, r, conn, host, db, t, k, v)
	} else if action == "GOTO" && db != "" && t != "" && n != "" {
		dumpIt(w, r, conn, host, db, t, o, d, n, g, k, v)
	} else if action == "BACK" {
		dumpIt(w, r, conn, host, db, "", "", "", "", "", "", "")
	} else if action != "" {
		shipMessage(w, host, db, "Unknown action: "+action)
	} else {
		dumpIt(w, r, conn, host, db, t, o, d, n, g, k, v)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	log.Println("[GET]", r.URL)
	user := q.Get("user") // Login via bookmark
	pass := q.Get("pass")
	host := q.Get("host")
	port := q.Get("port")
	dbms := q.Get("dbms")
	db :=   q.Get("db")
	if user != "" && pass != "" {
        log.Println("[LOGIN]",dbms,user,host, port,db)
		if dbms == "" {
			dbms = "mysql"
		}
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "3306"
		}
		setCredentials(w, r, dbms, host, port, user, pass)
		conn, err := sql.Open(dbms, dsn(user, pass, host, port, db))
		checkY(err)
		workload(w, r, conn, host)
	} else if dbms, host, port, user, pass, err := getCredentials(r); err == nil {
		conn, err := sql.Open(dbms, dsn(user, pass, host, port, db))
		checkY(err)
		workload(w, r, conn, host)
	} else {
		loginPageHandler(w, r)
	}
}

var EXPERTFLAG bool
var QUIETFLAG bool  // TODO: better located in session
var INFOFLAG bool
var DEBUGFLAG bool
var MODIFYFLAG bool
var READONLY bool
var CSS_FILE string
var TLS_PATH string


func main() {

	var SECURE = flag.String("s", "", "https Connection TLS")
	var HOST = flag.String("h", "localhost", "server name")
	var PORT = flag.Int("p", 8080, "server port")
	var INFO = flag.Bool("i", false, "include INFORMATION_SCHEMA in overview")
	var MODIFY = flag.Bool("m", false, "modify database schema: create, alter, drop tables (TODO)")
	var EXPERT = flag.Bool("x", false, "expert mode to access privileges, routines, triggers, views (TODO)")
	var QUIET = flag.Bool("q", false, "suppress messages (sql statements, row numbers of results)")
	var CSS = flag.String("c", "", "supply customized style in CSS file")
	var DEBUG = flag.Bool("d", false, "dynamically load  templates and css (DEBUG)")
	var READONLYFLAG = flag.Bool("r", false, "read-only access")

	flag.Parse()
	READONLY = *READONLYFLAG
	INFOFLAG = *INFO
	DEBUGFLAG = *DEBUG
	EXPERTFLAG = *EXPERT
	QUIETFLAG = *QUIET
	MODIFYFLAG = *MODIFY
	CSS_FILE = *CSS
	TLS_PATH = *SECURE
	initTemplate()

	portstring := ":" + strconv.Itoa(*PORT)
	var err error

	if CSS_FILE != "" && troubleF(CSS_FILE) == nil {
		http.HandleFunc("/"+CSS_FILE, cssHandler)
	}
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/test", testHandler)
	http.HandleFunc("/", indexHandler)

	if DEBUGFLAG {
		fmt.Println("dynamically loading  templates and css (DEBUG)")
	}
	if CSS_FILE != "" {
		fmt.Println("using style in " + CSS_FILE)
	}

	if !QUIETFLAG {
		fmt.Println("expert mode: showing sql statements")
	}

	if MODIFYFLAG {
		fmt.Println("modification of database schema enabled (TODO)")
	}

	if TLS_PATH != "" {
		if !strings.HasSuffix(TLS_PATH,"/") {
			TLS_PATH = TLS_PATH + "/"
		}
		certfile := TLS_PATH+"cert.pem"
		keyfile := TLS_PATH+"key.pem"
		if troubleF(certfile) != nil || troubleF(keyfile) != nil {
			fmt.Println("generating " + certfile + " and " + keyfile + " ...")
			generate_cert(*HOST, 2048, false, TLS_PATH)
		}
		fmt.Println("listening at https://" + *HOST + portstring + " using " + certfile + " and " + keyfile)
		err = http.ListenAndServeTLS(portstring, certfile, keyfile, nil)
	} else {
		fmt.Println("listening at http://" + *HOST + portstring)
		err = http.ListenAndServe(portstring, nil)
	}
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
