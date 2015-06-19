package main

import (
	"database/sql"
	"net/http"
	"net/url"
	"strconv"
)

// Dump all fields of a record, one column per line

func dumpFields(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, n string, nint int64, stmt sqlstring, v url.Values) {

	rows, err, _ := getRows(conn, stmt)
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
	for i, c := range cols {
		nv := getNullString(vals[i])
		row := []Entry{	escape(strconv.Itoa(i+1), ""),
						escape(c, ""),
						makeEntry(nv, db, t, c, "")}
		records = append(records, row)
	}

	v.Set("db", db)
	v.Set("t", t)

	left := Int64toa(maxInt64(nint-1, 1))
	var right string
	if rows.Next() {
		right = Int64toa(nint + 1)
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
	v.Set("n", n)

	var menu []Entry
	m := makeFreshQuery(db,t,o,d)
	m.Set("n",n)
	menu = append(menu,makeMenu(m, "action", "SELECTFORM","?"))
	menu = append(menu,makeMenu(m, "action", "INSERTFORM","+"))
	menu = append(menu,makeMenu(m, "action", "INFO","?"))

	tableOutFields(w, conn, host, db, t, "", o, d, "", n, "#", linkleft, linkright, head, records, menu)
}

func dumpKeyValue(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, k string, v string,  stmt sqlstring) {

	rows, err, _ := getRows(conn, stmt)
	checkY(err)
	defer rows.Close()

	primary := getPrimary(conn, t)
	head := []Entry{escape("#"), escape("Column"), escape("Data")}
	records := [][]Entry{}

	rows.Next()
	cols, vals, err := getRowScan(rows)
	for i, c := range cols {
		nv := getNullString(vals[i])
		row := []Entry{	escape(strconv.Itoa(i+1), ""),
						escape(c, ""),
						makeEntry(nv, db, t, c, "")}
		records = append(records, row)
	}

	q := makeFreshQuery(db,t,"","")
	q.Set("k",k)
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

	m := makeFreshQuery(db,t,"","")
	m.Set("k",k)
	m.Set("v",v)
	var menu []Entry
	menu = append(menu,makeMenu(m, "action", "SELECTFORM","?"))
	menu = append(menu,makeMenu(m, "action", "INSERTFORM","+"))
	menu = append(menu,makeMenu(m, "action", "KV_UPDATEFORM","~"))
	menu = append(menu,makeMenu(m, "action", "KV_DELETE","-"))
	menu = append(menu,makeMenu(m, "action", "INFO","?"))

	tableOutFields(w, conn, host, db, t, primary, k, "", k, v, k + " (ID) =", linkleft, linkright, head, records, menu)
}

