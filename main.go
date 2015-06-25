package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"regexp"
)
   
// HACK! TODO use proper accessors
func getDSN(conn *sql.DB) string {
	dsn := regexp.MustCompile("dsn:\"([^ ]*)\",").FindStringSubmatch(fmt.Sprintf("%#v",conn))
	
	if len(dsn)>1 {
		return dsn[1]
	} else {
		return ""
	}
}

func getHostDB(dsn string) (string,string) {
	db :=  regexp.MustCompile("/(.*)$").FindStringSubmatch(dsn)
	host := regexp.MustCompile("\\(([^ ]*):").FindStringSubmatch(dsn)
	if len(host)>1 && len(db) > 1{
		return host[1], db[1]
	} else {
		return "",""
	}
}


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

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if dbms, host, port, user, pass, base, err := getCredentialsFromRequest(r); err == nil {
		startStatus(w, r, dbms, host, port, user, pass, base, true)
	} else if dbms, host, port, user, pass, base, err := getCredentialsFromCookie(r); err == nil {
		startStatus(w, r, dbms, host, port, user, pass, base, false)
	} else {
		loginShowPage(w, r)
	}
}

func startStatus(w http.ResponseWriter, r *http.Request, 
	dbms string, user string, pass string, host string, port string, base string, new bool) {

	conn,err := getDB(w,r, dbms, user, pass, host, port, base, new)
	if err == nil {
		defer conn.Close()
		t, o, d, _, g, _, v := readRequest(r)
		showTableStatus(w, conn, host, base, t, o, d, g, v)
	} else {
		conn.Close()
		log.Println("[AUTH]",err)
		shipFatal(w,err)
	}
}


func loginShowPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, loginPage)
}


func workRouter(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string) {

	t, o, d, n, g, k, v := readRequest(r)

	q := r.URL.Query()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if q.Get("action") != "" && db != "" && t != "" {
		actionRouter(w, r, conn, host, db)
	} else if db != "" && t == "" {
		showTables(w, conn, host, db, t, o, d, g, v)
	} else if db != ""{
		dumpRouter(w, r, conn, t, o, d, n, g, k, v)
	} else {
		loginShowPage(w,r)
	}
}


func startWork(w http.ResponseWriter, r *http.Request, 
	dbms string, user string, pass string, host string, port string, base string, new bool) {
	conn,err := getDB(w,r, dbms, user, pass, host, port, base, new)
	if err == nil {
		defer conn.Close()
		workRouter(w, r, conn, host, base)
	} else {
		conn.Close()
		log.Println("[AUTH]",err)
		shipFatal(w,err)
	}
}


func getDB(w http.ResponseWriter, r *http.Request, 
	dbms string, user string, pass string, host string, port string, base string, new bool) (*sql.DB, error) {
	
	var conn *sql.DB
	var err error
	if strings.ToUpper(base)=="INFORMATION_SCHEMA" && !INFOFLAG {
		clearCredentials(w)
		err = fmt.Errorf("Access denied for user '"+user+"'@'"+host+"' to database '"+base+"'")
	} else {
		conn, err = sql.Open(dbms, dsn(user, pass, host, port, base))
		if err != nil {
			clearCredentials(w)
		} else {
			err = conn.Ping()
			if err != nil {
				clearCredentials(w)
			} else {
				if new {
					log.Println("[AUTH]", "Query", dbms, user, host, port, base)
					setCredentials(w, r, dbms, user, pass, host, port, base)
				}
			}
		}
	}
	return conn, err	
}


func indexHandler(w http.ResponseWriter, r *http.Request) {
	if dbms, user, pass, host, port, base, err := getCredentialsFromRequest(r); err == nil {
		startWork(w, r, dbms, user, pass, host, port, base, true)
	} else if dbms, user, pass, host, port, base, err := getCredentialsFromCookie(r); err == nil {
		startWork(w, r, dbms, user, pass, host, port, base, false)
	} else {
		loginShowPage(w, r)
	}
}

var EXPERTFLAG bool
var QUIETFLAG bool // TODO: better located in session
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
	var INFO = flag.Bool("i", false, "enable access to INFORMATION_SCHEMA")
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
	http.HandleFunc("/login", loginFormHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/test", testHandler)
	http.HandleFunc("/status", statusHandler)
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
		if !strings.HasSuffix(TLS_PATH, "/") {
			TLS_PATH = TLS_PATH + "/"
		}
		certfile := TLS_PATH + "cert.pem"
		keyfile := TLS_PATH + "key.pem"
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
