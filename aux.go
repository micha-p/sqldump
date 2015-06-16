package main

import (
	"log"
	"os"
)

/* 	Any identifier with double quotes in its name has to be escaped.
 */

func escape(name string, link ...string) Entry {
	if len(link) > 0 {
		return Entry{Text: name, Link: "/?" + link[0], Null: ""}
	} else {
		return Entry{Text: name, Link: "", Null: ""}
	}
}

func escapeNull() Entry {
	return Entry{Text: "", Link: "", Null: "NULL"}
}

func troubleF(filename string) error {
	_, err := os.Stat(filename)
	return err
}

// simple error checker
func checkY(err error) {
	if err != nil {
		log.Println("[ERROR]", err)
		os.Exit(1)
	}
}

// Compose dataSourceName from components and globals
// https://github.com/go-sql-driver/mysql/#dsn-data-source-name
func dsn(user string, pw string, host string, port string, db string) string {
	if DEBUGFLAG {
		log.Println("[DSN]", user+"@tcp("+host+":"+port+")/"+db)
	}
	return user + ":" + pw + "@tcp(" + host + ":" + port + ")/" + db
}

func maxI(a int, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

func minI(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
