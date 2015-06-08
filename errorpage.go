package main

import (
	"fmt"
	"net/http"
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

func shipErrorPage(w http.ResponseWriter, cred Access, db string, t string, cols []CContext) {

	c := EContext{
		CSS:      CSS_FILE,
		Action:   "BACK",
		Button:   "Back",
		Database: db,
		Table:    t,
		Back:     makeBack(cred.Host, db, t, "", "", ""),
		Columns:  cols,
		Trail:    makeTrail(cred.Host, db, t, "", "", "", ""),
	}

	if DEBUGFLAG {
		initTemplate()
	}
	err := templateError.Execute(w, c)
	checkY(err)
}

func checkErrorPage(w http.ResponseWriter, cred Access, db string, t string, query string, err error) {
	if err != nil {
		cols := []CContext{CContext{"1", "", "Query", "", "", query, ""}, CContext{"2", "", "Error", "", "", fmt.Sprint(err), ""}}
		shipErrorPage(w, cred, db, t, cols)
	}
}

func shipMessage(w http.ResponseWriter, cred Access, db string, msg string) {
	cols := []CContext{CContext{"1", "", "Message", "", "", msg, ""}}
	shipErrorPage(w, cred, db, "", cols)
}
