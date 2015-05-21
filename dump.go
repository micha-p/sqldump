package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
)

func dumpIt(w http.ResponseWriter, r *http.Request, cred Access) {
	q := r.URL.Query()
	db := q.Get("db")
	t := q.Get("t")
	o := q.Get("o")
	n := q.Get("n")

	v := url.Values{}
	trail := []Entry{}
	trail = append(trail, Entry{"/", cred.Host})

	if db == "" {
		q.Del("t")
		q.Del("o")
		q.Del("n")
		dumpHome(w, r, cred, trail, "/logout")
		return
	} else {
		v.Add("db", db)
		trail = append(trail, Entry{Link: "?" + v.Encode(), Label: db})
	}

	if t == "" {
		q.Del("o")
		q.Del("n")
		dumpTables(w, r, cred, trail, db, "?"+v.Encode())
		return
	} else {
		v.Add("t", t)
		trail = append(trail, Entry{Link: "/?" + v.Encode(), Label: t})
	}

	if n == "" {
		if o == "" {
			dumpRecords(w, r, cred, trail, db, t, o, "?"+q.Encode())
			return
		} else {
			v.Add("o", o)
			trail = append(trail, Entry{Link: "/?" + v.Encode(), Label: o})
			dumpOrdered(w, r, cred, trail, db, t, o, "?"+q.Encode())
			return
		}
	}

	if n != "" {
		if o == "" {
			dumpOne(w, r, cred, trail, db, t, o, n, "?"+q.Encode())
			return
		} else {
			v.Add("o", o)
			trail = append(trail, Entry{Link: "/?" + v.Encode(), Label: o})
			dumpOneOrdered(w, r, cred, trail, db, t, o, n, "?"+q.Encode())
			return
		}
	}
}

// Shows selection of databases at top level
func dumpHome(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, back string) {

	rows := getRows(cred, "", "show databases")
	defer rows.Close()

	menu := []Entry{}
	//menu = append(menu, Entry{"/logout", "Q"})
	records := [][]string{}
	head := []string{"Database"}
	var n int = 1
	for rows.Next() {
		var field string
		rows.Scan(&field)
		if EXPERTFLAG || INFOFLAG || field != "information_schema" {
			row := []string{href(r.URL.Host+"?"+"db="+field, strconv.Itoa(n)), field}
			records = append(records, row)
			n = n + 1
		}
	}
	tableOut(w, r, cred, back, head, records, trail, menu)
}

//  Dump all tables of a database
func dumpTables(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, db string, back string) {

	rows := getRows(cred, db, "show tables")
	defer rows.Close()

	menu := []Entry{}
	//menu = append(menu, Entry{"/logout", "Q"})
	records := [][]string{}
	head := []string{"Table", "Rows"}

	var n int = 1
	for rows.Next() {
		var field string
		var row []string
		var nrows string
		rows.Scan(&field)
		if db == "information_schema" {
			nrows = "?"
		} else {
			nrows = getCount(cred, db, field)
		}
		row = []string{href(r.URL.Host+"?"+r.URL.RawQuery+"&t="+field, strconv.Itoa(n)), field, nrows}
		records = append(records, row)
		n = n + 1
	}
	tableOut(w, r, cred, back, head, records, trail, menu)
}

//  Dump all records of a table, one per row
func dumpRecords(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, db string, t string, o string, back string) {
	dumpRows(w, r, cred, trail, db, t, o, back, "select * from "+template.HTMLEscapeString(t))
}

//  Dump all records of a table, one per row, ordered by one column
func dumpOrdered(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, db string, t string, o string, back string) {
	dumpRows(w, r, cred, trail, db, t, o, back, "select * from "+template.HTMLEscapeString(t)+" order by "+template.HTMLEscapeString(o))
}

//  Dump one record of a table
func dumpOne(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, db string, t string, o string, n string, back string) {
	dumpFields(w, r, cred, trail, db, t, o, n, back, "select * from "+template.HTMLEscapeString(t))
}

//  Dump one record of a table, ordered by one column
func dumpOneOrdered(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, db string, t string, o string, n string, back string) {
	dumpFields(w, r, cred, trail, db, t, o, n, back, "select * from "+template.HTMLEscapeString(t)+" order by "+template.HTMLEscapeString(o))
}

// http://stackoverflow.com/questions/17845619/how-to-call-the-scan-variadic-function-in-golang-using-reflection/17885636#17885636
// http://blog.golang.org/laws-of-reflection

func dumpValue(val interface{}) string {

	var r string
	b, ok := val.([]byte)

	if b != nil {
		if ok {
			r = string(b)
		} else {
			r = fmt.Sprint(val)
		}
	} else {
		r = "NIL"
	}
	return r
}

func dumpRows(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, db string, t string, o string, back string, query string) {

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("action", "add")
	linkinsert := "/?" + q.Encode()
	q.Set("action", "subset")
	linkselect := "/?" + q.Encode()
	q.Set("action", "show")
	linkshow := "?" + q.Encode()
	q.Del("action")
	menu := []Entry{}
	menu = append(menu, Entry{linkselect, "?"})
	menu = append(menu, Entry{linkshow, "i"})
	menu = append(menu, Entry{linkinsert, "+"})

	rows := getRows(cred, db, query)
	defer rows.Close()
	columns, err := rows.Columns()
	checkY(err)
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for i, _ := range columns {
		valuePtrs[i] = &values[i]
	}

	head := []string{}
	records := [][]string{}
	for _, title := range columns {
		q.Set("o", title)
		head = append(head, href("?"+q.Encode(), title))
	}
	q.Del("o")

	var n int = 1
	for rows.Next() {

		if o != "" {
			q.Set("o", o)
		}
		q.Set("n", strconv.Itoa(n))
		row := []string{href("?"+q.Encode(), strconv.Itoa(n))}

		err = rows.Scan(valuePtrs...)
		checkY(err)

		for i, _ := range columns {
			row = append(row, dumpValue(values[i]))
		}

		records = append(records, row)
		n = n + 1
	}
	tableOut(w, r, cred, back, head, records, trail, menu)
}

// Dump all fields of a record, one column per line
func dumpFields(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, db string, t string, o string, num string, back string, query string) {

	rows := getRows(cred, db, query)
	defer rows.Close()
	columns, err := rows.Columns()
	checkY(err)
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for i, _ := range columns {
		valuePtrs[i] = &values[i]
	}

	head := []string{"Column", "Data"}
	records := [][]string{}

	rec, err := strconv.Atoi(num)
	checkY(err)
	/*	nmax, err := strconv.Atoi(getCount(cred, db, t))
		checkY(err) */
	var n int = 1
rowLoop:
	for rows.Next() {

		// unfortunately we have to iterate up to row of interest
		if n == rec {
			err = rows.Scan(valuePtrs...)
			checkY(err)
			for i, _ := range columns {
				var row []string
				row = []string{strconv.Itoa(i + 1), columns[i], dumpValue(values[i])}
				records = append(records, row)
			}
			break rowLoop
		}
		n = n + 1
	}

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)

	q.Add("action", "add")
	linkinsert := "/?" + q.Encode()
	q.Set("action", "subset")
	//	linkselect := "/?" + q.Encode()
	q.Set("action", "show")
	linkshow := "?" + q.Encode()
	q.Del("action")
	/*	q.Set("n", strconv.Itoa(maxI(rec-1, 1)))
		linkleft := "?" + q.Encode()
		q.Set("n", strconv.Itoa(minI(rec+1, nmax)))
		linkright := "?" + q.Encode()*/

	menu := []Entry{}
	/*	menu = append(menu, Entry{linkleft, "<"})
		menu = append(menu, Entry{linkright, ">"})
	*/menu = append(menu, Entry{linkshow, "i"})
	menu = append(menu, Entry{linkinsert, "+"})
	//menu = append(menu, Entry{"/logout", "Q"})

	tableOutFields(w, r, cred, back, head, records, trail, menu)
}
