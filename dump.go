package main

import (
	"html/template"
	"net/http"
	"net/url"
	"strconv"
)


func dumpIt(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	db := q.Get("db")
	t := q.Get("t")
	n := q.Get("n")

	if db == "" {
		dumpHome(w, r, "/logout")
	} else if t == "" {
		q.Del("db")
		dumpTables(w, r, db, "?"+q.Encode())
	} else if n == "" {
		q.Del("t")
		dumpRecords(w, r, db, t, "?"+q.Encode())
	} else {
		q.Del("n")
		dumpFields(w, r, db, t, n, "?"+q.Encode())
	}
}

// Shows selection of databases at top level
func dumpHome(w http.ResponseWriter, r *http.Request, back string) {

	rows := getRows(r, "", "show databases")
	defer rows.Close()

	trail := []Entry{}	
	trail = append(trail, Entry{"/","root"})
	menu := []Entry{}
	menu = append(menu, Entry{"/logout","Q"})
	records := [][]string{}
	head := []string{"Database"}
	var n int = 1
	for rows.Next() {
		var field string
		rows.Scan(&field)
		row := []string{href(r.URL.Host+"?"+"db="+field, strconv.Itoa(n)), field}
		records = append(records, row)
		n = n + 1
	}
	tableOut(w, r, back, head, records, trail, menu)
}

//  Dump all tables of a database
func dumpTables(w http.ResponseWriter, r *http.Request, db string, back string) {

	rows := getRows(r, db, "show tables")
	defer rows.Close()

	trail := []Entry{}	
	trail = append(trail, Entry{"/","root"})
	trail = append(trail, Entry{"?db=" + db,db})
	menu := []Entry{}
	menu = append(menu, Entry{"/logout","Q"})
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
			nrows = getCount(r, db, field)
		}
		row = []string{href(r.URL.Host+"?"+r.URL.RawQuery+"&t="+field, strconv.Itoa(n)), field, nrows}
		records = append(records, row)
		n = n + 1
	}
	tableOut(w, r, back, head, records, trail, menu)
}

//  Dump all records of a table, one per row
func dumpRecords(w http.ResponseWriter, r *http.Request, db string, t string, back string) {
	dumpRows(w, r, db, t, back, "select * from "+template.HTMLEscapeString(t))
}

func dumpRows(w http.ResponseWriter, r *http.Request, db string, t string, back string, query string) {

	rows := getRows(r, db, query)
	defer rows.Close()
	cols, err := rows.Columns()
	checkY(err)

	trail := []Entry{}	
	trail = append(trail, Entry{"/","root"})
	
	q := url.Values{}
	q.Add("db", db)
	trail = append (trail, Entry{Link:"/?" + q.Encode() , Label: db,})
	q.Add("t", t)
	trail = append (trail, Entry{Link:"/?" + q.Encode() , Label: t,})


	/*  credits:
	 * 	http://stackoverflow.com/questions/19991541/dumping-mysql-tables-to-json-with-golang
	 * 	http://go-database-sql.org/varcols.html
	 */
	raw := make([]interface{}, len(cols))
	val := make([]interface{}, len(cols))

	for i := range val {
		raw[i] = &val[i]
	}

	head := []string{}
	for _, column := range cols {
		head = append(head, column)
	}
		
	records := [][]string{}
	var n int = 1
	for rows.Next() {

		q.Set("n",strconv.Itoa(n))
		row := []string{href("?" + q.Encode(), strconv.Itoa(n))}

		err = rows.Scan(raw...)
		checkY(err)

		for _, col := range val {
			if col != nil {
				row = append(row, string(col.([]byte)))
			}
		}
		records = append(records, row)
		n = n + 1
	}
	
	q.Add("action", "add")
	linkinsert := "/?" + q.Encode()
	q.Set("action", "subset")
	linkselect := "/?" + q.Encode()

	menu := []Entry{}
	menu = append(menu, Entry{linkselect,"?"})
	menu = append(menu, Entry{linkinsert,"+"})
	menu = append(menu, Entry{"/logout","Q"})

	
	tableOut(w, r, back, head, records, trail, menu)
}

// Dump all fields of a record, one column per line
func dumpFields(w http.ResponseWriter, r *http.Request, db string, t string, num string, back string) {

	rows := getRows(r, db, "select * from "+template.HTMLEscapeString(t))
	defer rows.Close()

	columns, err := rows.Columns()
	checkY(err)

	trail := []Entry{}	
	trail = append(trail, Entry{"/","root"})
	
	q := url.Values{}
	q.Add("db", db)
	trail = append (trail, Entry{Link:"/?" + q.Encode() , Label: db,})
	q.Add("t", t)
	trail = append (trail, Entry{Link:"/?" + q.Encode() , Label: t,})
	q.Add("n", num)
	trail = append (trail, Entry{Link:"/?" + q.Encode() , Label: num,})

	raw := make([]interface{}, len(columns))
	val := make([]interface{}, len(columns))

	for i := range val {
		raw[i] = &val[i]
	}

    head := []string{"Column", "Data"}
	records := [][]string{}

	rec, err := strconv.Atoi(num)
	checkY(err)	
	nmax, err := strconv.Atoi(getCount(r, db, t))
	checkY(err)	
	var n int = 1
rowLoop:
	for rows.Next() {

		// unfortunately we have to iterate up to row of interest
		if n == rec {
			err = rows.Scan(raw...)
			checkY(err)

			for i, col := range val {
				var row []string
				if col != nil {
					row = []string{strconv.Itoa(i), columns[i], string(col.([]byte))}
					records = append(records, row)
				}
			}
			break rowLoop
		}
		n = n + 1
	}
	
	q.Set("n", strconv.Itoa(maxI(rec-1, 1)))
	linkleft := "?" + q.Encode()
	q.Set("n", strconv.Itoa(minI(rec+1, nmax)))
	linkright := "?" + q.Encode()
	
	q.Add("action", "add")
	linkinsert := "/?" + q.Encode()
	q.Set("action", "subset")
//	linkselect := "/?" + q.Encode()
	q.Set("action", "show")
	linkshow := "?" + q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, Entry{linkleft,"<"})
	menu = append(menu, Entry{linkright,">"})
	menu = append(menu, Entry{linkshow,"?"})
	menu = append(menu, Entry{linkinsert,"+"})
	menu = append(menu, Entry{"/logout","Q"})
	
	tableOutFields(w, r, back, head, records, trail, menu)
}
