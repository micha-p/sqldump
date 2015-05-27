package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
)

func dumpIt(w http.ResponseWriter, cred Access, db string, t string, o string, od string, n string) {

	v := url.Values{}
	trail := []Entry{}
	trail = append(trail, Entry{"/", cred.Host})
	var query string

	if db == "" {
		dumpHome(w, cred, trail, "/logout")
		return
	} else {
		v.Add("db", db)
		trail = append(trail, Entry{Link: "?" + v.Encode(), Label: db})
	}

	if t == "" {
		dumpTables(w, db, cred, trail, "?"+v.Encode())
		return
	} else {
		v.Add("t", t)
		trail = append(trail, Entry{Link: "/?" + v.Encode(), Label: t})
	}

	if n == "" {
		if o != "" {
			v.Add("o", o)
			trail = append(trail, Entry{Link: "/?" + v.Encode(), Label: o + "&uarr;"})
			query = "select * from "+template.HTMLEscapeString(t)+" order by "+template.HTMLEscapeString(o)
		} else if od != "" {
			v.Add("od", od)
			trail = append(trail, Entry{Link: "/?" + v.Encode(), Label: od + "&darr;"})
			query = "select * from "+template.HTMLEscapeString(t)+" order by "+template.HTMLEscapeString(od)+" desc"
		} else {
			query = "select * from "+template.HTMLEscapeString(t)
		}
		dumpRows(w, db, t, o, od, cred, trail, "?"+v.Encode(),query)
	} else {
		if o != "" {
			v.Add("o", o)
			trail = append(trail, Entry{Link: "/?" + v.Encode(), Label: o + "&uarr;"})
	        query = "select * from "+template.HTMLEscapeString(t)+" order by "+template.HTMLEscapeString(o)
		} else if od != "" {
			v.Add("od", od)
			trail = append(trail, Entry{Link: "/?" + v.Encode(), Label: od + "&darr;"})
	        query = "select * from "+template.HTMLEscapeString(t)+" order by "+template.HTMLEscapeString(o) + " desc"
		} else {
	        query = "select * from "+template.HTMLEscapeString(t)
		}
		dumpFields(w, db, t, o, od, n, cred, trail, "?"+v.Encode(), query)
	}
}

// Shows selection of databases at top level
func dumpHome(w http.ResponseWriter, cred Access, trail []Entry, back string) {

	q := url.Values{}
	rows := getRows(cred, "", "show databases")
	defer rows.Close()

	menu := []Entry{}
	records := [][]string{}
	head := []string{"Database"}
	var n int = 1
	for rows.Next() {
		var field string
		rows.Scan(&field)
		if EXPERTFLAG || INFOFLAG || field != "information_schema" {
			q.Set("db", field)
			link := "/?" + q.Encode()
			row := []string{href(link, strconv.Itoa(n)), href(link, field)}
			records = append(records, row)
			n = n + 1
		}
	}
	tableOut(w, cred, "", "", back, head, records, trail, menu)
}

//  Dump all tables of a database
func dumpTables(w http.ResponseWriter, db string, cred Access, trail []Entry, back string) {

	q := url.Values{}
	q.Add("db", db)
	rows := getRows(cred, db, "show tables")
	defer rows.Close()

	menu := []Entry{}
	records := [][]string{}
	head := []string{"Table", "Rows"}

	var n int = 1
	for rows.Next() {
		var field string
		var nrows string
		rows.Scan(&field)
		if db == "information_schema" {
			nrows = "?"
		} else {
			nrows = getCount(cred, db, field)
		}
		q.Set("t", field)
		link := "/?" + q.Encode()
		row := []string{href(link, strconv.Itoa(n)), href(link, field), nrows}
		records = append(records, row)
		n = n + 1
	}
	tableOut(w, cred, db, "", back, head, records, trail, menu)
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

func dumpRows(w http.ResponseWriter, db string, t string, o string, od string, cred Access, trail []Entry, back string, query string) {

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
		if o == title {
			q.Set("od", title)
			q.Del("o")
			head = append(head, href("?"+q.Encode(), title+"&uarr;"))
		} else if od == title {
			q.Set("o", title)
			q.Del("od")
			head = append(head, href("?"+q.Encode(), title+"&darr;"))
		} else {
			q.Set("o", title)
			head = append(head, href("?"+q.Encode(), title))
		}
	}
	q.Del("o")
	q.Del("od")
	if o != "" {
		q.Set("o", o)
	}
	if od != "" {
		q.Set("od", od)
	}
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
	tableOut(w, cred, db, t, back, head, records, trail, menu)
}

// Dump all fields of a record, one column per line
func dumpFields(w http.ResponseWriter, db string, t string, o string, od string, n string, cred Access, trail []Entry, back string, query string) {

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

	rec, err := strconv.Atoi(n)
	checkY(err)
	var iter int = 1
rowLoop:
	for rows.Next() {

		// unfortunately we have to iterate up to row of interest
		if iter == rec {
			err = rows.Scan(valuePtrs...)
			checkY(err)
			for i, _ := range columns {
				var row []string
				row = []string{strconv.Itoa(i + 1), columns[i], dumpValue(values[i])}
				records = append(records, row)
			}
			break rowLoop
		}
		iter = iter + 1
	}

	v := url.Values{}
	v.Add("db", db)
	v.Add("t", t)

	v.Add("action", "add")
	linkinsert := "/?" + v.Encode()
	v.Set("action", "show")
	linkshow := "?" + v.Encode()
	v.Del("action")

	menu := []Entry{}
	menu = append(menu, Entry{linkshow, "i"})
	menu = append(menu, Entry{linkinsert, "+"})

	nint, err := strconv.Atoi(n)
	nmax, err := strconv.Atoi(getCount(cred, db, t))
	left := strconv.Itoa(maxI(nint-1, 1))
	right := strconv.Itoa(minI(nint+1, nmax))

	if o != "" {
		v.Set("o", o)
	}
	if od != "" {
		v.Set("od", od)
	}
	v.Set("n", left)
	linkleft := "?" + v.Encode()
	v.Set("n", right)
	linkright := "?" + v.Encode()

	tableOutFields(w, cred, db, t, n, linkleft, linkright, back, head, records, trail, menu)
}
