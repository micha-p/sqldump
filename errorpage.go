package main

import (
	"fmt"
	"net/http"
	"database/sql"
)

type EContext struct {
	CSS      string
	Action   string
	Button   string
	Database string
	Table    string
	Back     string
	Columns  []CContext
	Trail    []Entry
}

func shipErrorPage(w http.ResponseWriter, conn *sql.DB, t string, cols []CContext) {

	host,db := getHostDB(getDSN(conn))
	c := EContext{
		CSS:      CSS_FILE,
		Action:   "BACK",
		Button:   "Back",
		Database: db,
		Table:    t,
		Back:     makeBack(host, db, t, "", "", ""),
		Columns:  cols,
		Trail:    makeTrail(host, db, t, "", "", [][]Clause{}),
	}

	if DEBUGFLAG {
		initTemplate()
	}
	err := templateError.Execute(w, c)
	checkY(err)
}

func checkErrorPage(w http.ResponseWriter, conn *sql.DB, t string, query sqlstring, err error) {
	if err != nil {
		s := sql2str(query)
		cols := []CContext{CContext{"1", "", "Query", "", "", "", "valid", s, ""},
			CContext{"2", "", "Error", "", "", "", "valid", fmt.Sprint(err), ""}}
		shipErrorPage(w, conn, t, cols)
	}
}

func shipMessage(w http.ResponseWriter, conn *sql.DB, msg string) {
	cols := []CContext{CContext{"1", "", "Message", "", "", "", "valid", msg, ""}}
	shipErrorPage(w, conn, "", cols)
}
