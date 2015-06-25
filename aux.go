package main

import (
	"log"
	"os"
	"strconv"
)


func escape(name string, arg ...string) Entry {
	if len(arg) == 2 && arg[1] !="" {
		return Entry{Text: name, Link: "/"+arg[0] + "?" + arg[1], Null: ""}
	} else if len(arg) == 2 && arg[1] =="" {
		return Entry{Text: name, Link: "/"+arg[0], Null: ""}
	} else if len(arg) == 2 && arg[0] =="" {
		return Entry{Text: name, Link: "/?"+arg[1], Null: ""}
	} else if len(arg) == 1 {
		return Entry{Text: name, Link: "/?" + arg[0], Null: ""}
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


func Atoi64(s string) (int64, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	return n, err
}

func Int64toa(i int64) string {
	return strconv.FormatInt(i, 10)
}

func maxInt(a int, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

func minInt(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func maxInt64(a int64, b int64) int64 {
	if a >= b {
		return a
	} else {
		return b
	}
}

func minInt64(a int64, b int64) int64 {
	if a < b {
		return a
	} else {
		return b
	}
}
