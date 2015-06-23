package main

import (
	"database/sql"
	"net/http"
	"net/url"
	"strconv"
)

func showFields(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, n string, nint int64, nmax int64, stmt sqlstring, q url.Values) {

	whereStack := WhereQuery2Stack(q, getColumnInfo(conn, t))
	q.Set("db", db)
	q.Set("t", t)

	left := Int64toa(maxInt64(nint-1, 1))
	var right string
	right = Int64toa(minInt64(nint+1, nmax))
	if o != "" {
		q.Set("o", o)
	}
	if d != "" {
		q.Set("d", d)
	}
	q.Set("n", left)
	linkleft := escape("<", q.Encode())
	q.Set("n", right)
	linkright := escape(">", q.Encode())
	q.Set("n", n)
	nstring := Int64toa(nint)

	menu := makeMenu3(q)
	verticalView(w, conn, stmt, host, db, t, o, d /* primary: */, "", nstring, "#", linkleft, linkright, menu, whereStack, q)
}

func verticalView(w http.ResponseWriter, conn *sql.DB, stmt sqlstring, host string, db string, t string, o string, d string,
	primary string, counter string, counterlabel string,
	linkleft Entry, linkright Entry, menu []Entry, whereStack [][]Clause, q url.Values) {

	rows, err, _ := getRows(conn, stmt)
	defer rows.Close()
	rows.Next()
	cols, vals, err := getRowScan(rows)
	checkY(err)

	var title_column Entry
	if d == "" {
		q.Set("d", "1")
		title_column = escape("Column", q.Encode())
		q.Del("d")
	} else {
		count := len(cols)
		newcols := make([]string, count)
		n := 0
		for i := count - 1; i >= 0; i-- {
			newcols[n] = cols[i]
			n = n + 1
		}
		q.Del("d")
		cols = newcols
		title_column = escape(makeTitleWithArrow("Column", "", d), q.Encode()) // it is really an index
		q.Set("d", d)
	}
	home := url.Values{}
	home.Add("db", db)
	home.Add("t", t)
	head := []Entry{escape("#", home.Encode()), title_column, escape("Data")}

	records := [][]Entry{}
	for i, c := range cols {
		nv := getNullString(vals[i])
		label := c
		if c == primary {
			label = label + " (ID)"
		}
		row := []Entry{Entry{Text: strconv.Itoa(i + 1)},
			Entry{Text: label},
			makeEntry(nv, db, t, c, "",q)}
		records = append(records, row)
	}
	tableOutFields(w, conn, host, db, t, o, d, counter, counterlabel, linkleft, linkright, head, records, menu, whereStack)
}

func showKeyValue(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, k string, v string, stmt sqlstring) {

	q := makeFreshQuery(db, t, "", "")
	q.Set("k", k)
	next, err := getSingleValue(conn, sqlSelect(k, t)+sqlWhere(k, ">", v)+sqlOrder(k, "")+sqlLimit(1, 0))
	if err == nil {
		q.Set("v", next)
	} else {
		q.Set("v", v)
	}
	linkright := escape(">", q.Encode())
	prev, err := getSingleValue(conn, sqlSelect(k, t)+sqlWhere(k, "<", v)+sqlOrder(k, "1")+sqlLimit(1, 0))
	if err == nil {
		q.Set("v", prev)
	} else {
		q.Set("v", v)
	}
	linkleft := escape("<", q.Encode())

	m := makeFreshQuery(db, t, "", "")
	m.Set("k", k)
	m.Set("v", v)
	menu := makeMenu5(m)
	verticalView(w, conn, stmt, host, db, t /* order: */, k, d, k, v, k+" (ID) =", linkleft, linkright, menu, [][]Clause{}, q)
}
