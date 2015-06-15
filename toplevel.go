package main

import (
	"net/http"
	"net/url"
	"strconv"
)

func dumpIt(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, o string, d string, n string, k string, v string) {

	if db == "" {
		dumpHome(w, cred)
		return
	} else if t == "" {
		dumpTables(w, db, cred)
	} else {
		dumpSelection(w, cred, db, t, o, d, n, k, v)
	}
}

// Shows selection of databases at top level
func dumpHome(w http.ResponseWriter, cred Access) {

	q := url.Values{}
	rows, err := getRows(cred, "", "show databases")
	checkY(err)
	defer rows.Close()

	records := [][]Entry{}
	head := []Entry{{"#", ""}, {"Database", ""}}
	var n int = 1
	for rows.Next() {
		var field string
		rows.Scan(&field)
		if EXPERTFLAG || INFOFLAG || field != "information_schema" {
			q.Set("db", field)
			link := q.Encode()
			row := []Entry{escape(strconv.Itoa(n), link), escape(field, link)}
			records = append(records, row)
			n = n + 1
		}
	}
	tableOutSimple(w, cred, "", "", head, records, []Entry{})
}

//  Dump all tables of a database
func dumpTables(w http.ResponseWriter, db string, cred Access) {

	q := url.Values{}
	q.Add("db", db)
	rows, err := getRows(cred, db, "show tables")
	checkY(err)
	defer rows.Close()

	records := [][]Entry{}
	head := []Entry{{"#", ""}, {"Table", ""}, {"Rows", ""}}

	var n int = 1
	for rows.Next() {
		var field string
		var nrows string
		rows.Scan(&field)
		nrows = getCount(cred, db, field)

		q.Set("t", field)
		link := q.Encode()
		row := []Entry{escape(strconv.Itoa(n), link), escape(field, link), escape(nrows, "")}
		records = append(records, row)
		n = n + 1
	}
	tableOutSimple(w, cred, db, "", head, records, []Entry{})
}
