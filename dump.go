package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
)


    
func dumpIt(w http.ResponseWriter, r *http.Request, cred Access) {

	db,t,o,od,n := readRequest(r)
	q := r.URL.Query()
	v := url.Values{}
	trail := []Entry{}
	trail = append(trail, Entry{"/", cred.Host})

	if db == "" {
		q.Del("t")
		q.Del("o")
		q.Del("od")
		q.Del("n")
		dumpHome(w, r, cred, trail, "/logout")
		return
	} else {
		v.Add("db", db)
		trail = append(trail, Entry{Link: "?" + v.Encode(), Label: db})
	}

	if t == "" {
		q.Del("o")
		q.Del("od")
		q.Del("n")
		dumpTables(w, r, cred, trail, "?"+v.Encode())
		return
	} else {
		v.Add("t", t)
		trail = append(trail, Entry{Link: "/?" + v.Encode(), Label: t})
	}

	if n == "" {
		if o != "" {
			v.Add("o", o)
			trail = append(trail, Entry{Link: "/?" + v.Encode(), Label: o + "&uarr;"})
			dumpOrdered(w, r, cred, trail, "?"+q.Encode())
			return
		} else if od != ""{
			v.Add("od", od)
			trail = append(trail, Entry{Link: "/?" + v.Encode(), Label: od + "&darr;"})
			dumpOrderedDesc(w, r, cred, trail, "?"+q.Encode())
			return
		} else {
			dumpRecords(w, r, cred, trail, "?"+q.Encode())
			return
		}
	} else {
		if o != "" {
			v.Add("o", o)
			trail = append(trail, Entry{Link: "/?" + v.Encode(), Label: o + "&uarr;"})
			dumpOneOrdered(w, r, cred, trail, "?"+q.Encode())
			return
		} else if od != ""{
			v.Add("od", od)
			trail = append(trail, Entry{Link: "/?" + v.Encode(), Label: od + "&darr;"})
			dumpOneOrderedDesc(w, r, cred, trail, "?"+q.Encode())
			return
		} else {
			dumpOne(w, r, cred, trail, "?"+q.Encode())
			return
		}
	}
}

// Shows selection of databases at top level
func dumpHome(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, back string) {

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
			link := r.URL.Host+"?"+"db="+field
			row := []string{href(link, strconv.Itoa(n)), href(link,field)}
			records = append(records, row)
			n = n + 1
		}
	}
	tableOut(w, r, cred, back, head, records, trail, menu)
}

//  Dump all tables of a database
func dumpTables(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, back string) {

	db,_,_,_,_ := readRequest(r)
	rows := getRows(cred, db, "show tables")
	defer rows.Close()

	menu := []Entry{}
	//menu = append(menu, Entry{"/logout", "Q"})
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
		link := r.URL.Host+"?"+r.URL.RawQuery+"&t="+field
		row := []string{href(link, strconv.Itoa(n)), href(link, field), nrows}
		records = append(records, row)
		n = n + 1
	}
	tableOut(w, r, cred, back, head, records, trail, menu)
}

//  Dump all records of a table, one per row
func dumpRecords(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, back string) {
	_,t,_,_,_ := readRequest(r)
	dumpRows(w, r, cred, trail, back, "select * from "+template.HTMLEscapeString(t))
}

//  Dump all records of a table, one per row, ordered by one column
func dumpOrdered(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, back string) {
	_,t,o,_,_ := readRequest(r)
	dumpRows(w, r, cred, trail, back, "select * from "+template.HTMLEscapeString(t)+" order by "+template.HTMLEscapeString(o))
}

//  Dump all records of a table, one per row, ordered by one column DESC
func dumpOrderedDesc(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, back string) {
	_,t,_,od,_ := readRequest(r)
	dumpRows(w, r, cred, trail, back, "select * from "+template.HTMLEscapeString(t)+" order by "+template.HTMLEscapeString(od) +" desc")
}

//  Dump one record of a table
func dumpOne(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, back string) {
	_,t,_,_,_ := readRequest(r)
	dumpFields(w, r, cred, trail, back, "select * from "+template.HTMLEscapeString(t))
}

//  Dump one record of a table, ordered by one column
func dumpOneOrdered(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, back string) {
	_,t,o,_,_ := readRequest(r)
	dumpFields(w, r, cred, trail, back, "select * from "+template.HTMLEscapeString(t)+" order by "+template.HTMLEscapeString(o))
}

//  Dump one record of a table, ordered by one column DESC
func dumpOneOrderedDesc(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, back string) {
	_,t,_,od,_ := readRequest(r)
	dumpFields(w, r, cred, trail, back, "select * from "+template.HTMLEscapeString(t)+" order by "+template.HTMLEscapeString(od)+" desc")
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

func dumpRows(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, back string, query string) {

	db,t,o,od,_ := readRequest(r)
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
			head = append(head, href("?"+q.Encode(), title + "&uarr;"))
		} else if od == title {
			q.Set("o", title)
			q.Del("od")
			head = append(head, href("?"+q.Encode(), title + "&darr;"))
		} else {
			q.Set("o", title)
			head = append(head, href("?"+q.Encode(), title))
		}
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
func dumpFields(w http.ResponseWriter, r *http.Request, cred Access, trail []Entry, back string, query string) {

	db,t,_,_,n := readRequest(r)
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

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)

	q.Add("action", "add")
	linkinsert := "/?" + q.Encode()
	q.Set("action", "show")
	linkshow := "?" + q.Encode()
	q.Del("action")

	menu := []Entry{}
    menu = append(menu, Entry{linkshow, "i"})
	menu = append(menu, Entry{linkinsert, "+"})

	tableOutFields(w, r, cred, back, head, records, trail, menu)
}
