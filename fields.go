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
	vmap := getNullStringMap(rows)

	home := url.Values{}
	home.Add("db", db)
	home.Add("t", t)
	head := []Entry{escape("#", home.Encode()), escape("Column"), escape("Data")}
	records := [][]Entry{}

	i := 1
	for f, nv := range vmap { // TODO should be range cols
		v := nv.String
		var row []Entry
		row = []Entry{escape(strconv.Itoa(i), ""), escape(f, ""), escape(v, "")}
		records = append(records, row)
		i = i + 1
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

	tableOutFields(w, conn, host, db, t, "", o, d, "", n, linkleft, linkright, head, records, menu)
}

func dumpKeyValue(w http.ResponseWriter, db string, t string, k string, v string, conn *sql.DB, host string, query sqlstring) {

	rows, err := getRows(conn, query)
	checkY(err)
	vmap := getNullStringMap(rows)
	primary := getPrimary(conn, t)
	head := []Entry{escape("#"), escape("Column"), escape("Data")}
	records := [][]Entry{}

	i := 1
	for f, nv := range vmap { // TODO should be range cols
		v := nv.String
		var row []Entry
		row = []Entry{escape(strconv.Itoa(i), ""), escape(f, ""), escape(v, "")}
		records = append(records, row)
		i = i + 1
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
	tableOutFields(w, conn, host, db, t, primary, k, "", k, v, linkleft, linkright, head, records, menu)
}
