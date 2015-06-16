package main

import (
	"database/sql"
	"net/http"
	"net/url"
	"strconv"
)

func dumpIt(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, o string, d string, n string, k string, v string) {

	if db == "" {
		dumpHome(w, conn, host)
		return
	} else if t == "" {
		dumpTables(w, db, conn, host)
	} else if k != "" && v != "" && k == getPrimary(conn, host, db, t) {
		dumpKeyValue(w, db, t, k, v, conn, host, sqlStar(t) + sqlWhere(k,"=",v))
	} else{
		dumpSelection(w, r, conn, host, db, t, o, d, n, k, v)
	}
}

// Shows selection of databases at top level
func dumpHome(w http.ResponseWriter, conn *sql.DB, host string) {

	q := url.Values{}
	rows, err := getRows(conn, host, "", "SHOW DATABASES")
	checkY(err)
	defer rows.Close()

	records := [][]Entry{}
	head := []Entry{{"#", "", ""}, {"Database", "", ""}}
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
	tableOutSimple(w, conn, host, "", "", head, records, []Entry{})
}

//  Dump all tables of a database
func dumpTables(w http.ResponseWriter, db string, conn *sql.DB, host string) {

	q := url.Values{}
	q.Add("db", db)
	rows, err := getRows(conn, host, db, "SHOW TABLES")
	checkY(err)
	defer rows.Close()

	records := [][]Entry{}
	head := []Entry{{"#", "", ""}, {"Table", "", ""}, {"Rows", "", ""}}

	var n int = 1
	for rows.Next() {
		var field string
		var nrows string
		rows.Scan(&field)
		nrows = getCount(conn, host, db, field)

		q.Set("t", field)
		link := q.Encode()
		row := []Entry{escape(strconv.Itoa(n), link), escape(field, link), escape(nrows, "")}
		records = append(records, row)
		n = n + 1
	}
	tableOutSimple(w, conn, host, db, "", head, records, []Entry{})
}
