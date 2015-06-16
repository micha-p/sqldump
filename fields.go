package main

import (
	"database/sql"
	"net/http"
	"net/url"
	"strconv"
)

// Dump all fields of a record, one column per line

func dumpFields(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, n string, nint int, query sqlstring, v url.Values) {

	rows, err := getRows(conn, query)
	defer rows.Close()
	checkY(err)
	
	home := url.Values{}
	home.Add("db", db)
	home.Add("t", t)
	head := []Entry{escape("#", home.Encode()), escape("Column"), escape("Data")}
	records := [][]Entry{}

	rows.Next()
	cols, vals, err := getRowScan(rows)	
	checkY(err)
	for i, f := range cols {
		nv := getNullString(vals[i])
		if nv.Valid {
			row := []Entry{escape(strconv.Itoa(i+1), ""), escape(f, ""), escape(nv.String, "")}
			records = append(records, row)
		} else {
			row := []Entry{escape(strconv.Itoa(i+1), ""), escape(f, ""), escapeNull()}
			records = append(records, row)
		}
	}

	v.Add("db", db)
	v.Add("t", t)
	v.Add("action", "ADD")
	linkinsert := v.Encode()
	v.Set("action", "INFO")
	linkinfo := v.Encode()
	v.Del("action")

	menu := []Entry{}
	menu = append(menu, escape("+", linkinsert))
	menu = append(menu, escape("i", linkinfo))

	left := strconv.Itoa(maxI(nint-1, 1))
	var right string
	if rows.Next() {
		right = strconv.Itoa(nint + 1)
	} else {
		right = n
	}
	if o != "" {
		v.Set("o", o)
	}
	if d != "" {
		v.Set("d", d)
	}
	v.Set("n", left)
	linkleft := escape("<", v.Encode())
	v.Set("n", right)
	linkright := escape(">", v.Encode())

	tableOutFields(w, conn, host, db, t, "", o, d, "", n, "#", linkleft, linkright, head, records, menu)
}

func dumpKeyValue(w http.ResponseWriter, db string, t string, k string, v string, conn *sql.DB, host string, query sqlstring) {

	rows, err := getRows(conn, query)
	checkY(err)
	defer rows.Close()
	
	primary := getPrimary(conn, t)
	head := []Entry{escape("#"), escape("Column"), escape("Data")}
	records := [][]Entry{}

	rows.Next()
	cols, vals, err := getRowScan(rows)
	for i, f := range cols {
		nv := getNullString(vals[i])
		if nv.Valid {
			row := []Entry{escape(strconv.Itoa(i+1), ""), escape(f, ""), escape(nv.String, "")}
			records = append(records, row)
		} else {
			row := []Entry{escape(strconv.Itoa(i+1), ""), escape(f, ""), escapeNull()}
			records = append(records, row)
		}
	}

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("action", "ADD")
	linkinsert := q.Encode()
	q.Set("action", "INFO")
	linkinfo := q.Encode()
	q.Add("k", k)
	q.Add("v", v)
	q.Set("action", "DELETEPRI")
	linkDELETEPRI := q.Encode()
	q.Set("action", "EDITFORM")
	linkedit := q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, escape("+", linkinsert))
	menu = append(menu, escape("~", linkedit))
	menu = append(menu, escape("-", linkDELETEPRI))
	menu = append(menu, escape("i", linkinfo))

	next, err := getSingleValue(conn, host, db, sqlSelect(k, t)+sqlWhere(k, ">", v)+sqlOrder(k, "")+sqlLimit(1, 0))
	if err == nil {
		q.Set("v", next)
	} else {
		q.Set("v", v)
	}
	linkright := escape(">", q.Encode())
	prev, err := getSingleValue(conn, host, db, sqlSelect(k, t)+sqlWhere(k, "<", v)+sqlOrder(k, "1")+sqlLimit(1, 0))
	if err == nil {
		q.Set("v", prev)
	} else {
		q.Set("v", v)
	}
	linkleft := escape("<", q.Encode())
	tableOutFields(w, conn, host, db, t, primary, k, "", k, v, k + " (ID) =", linkleft, linkright, head, records, menu)
}
