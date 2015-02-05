package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
)


func getCols(r *http.Request, database string, table string) []string {
	rows := getRows(r, database, "select * from "+template.HTMLEscapeString(table))
	defer rows.Close()

	cols, err := rows.Columns()
	checkY(err)
	return cols
}

func getRows(r *http.Request, database string, stmt string) *sql.Rows {
	user, pw, h, p := getCredentials(r)
	conn, err := sql.Open("mysql", dsn(user, pw, h, p, database))
	checkY(err)
	defer conn.Close()

	statement, err := conn.Prepare(stmt)
	checkY(err)
	rows, err := statement.Query()
	checkY(err)

	return rows
}

func getCount(r *http.Request, database string, table string) string {

	rows := getRows(r, database, "select count(*) from "+template.HTMLEscapeString(table))
	defer rows.Close()

	rows.Next()
	var field string
	rows.Scan(&field)
	return field
}

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

	var n int = 1
	records := [][]string{}
	head := []string{"Database"}
	for rows.Next() {
		var field string
		rows.Scan(&field)
		row := []string{href(r.URL.Host+"?"+"db="+field, strconv.Itoa(n)), field}
		records = append(records, row)
		n = n + 1
	}
	tableOut(w, r, back, head, records)
}

//  Dump all tables of a database
func dumpTables(w http.ResponseWriter, r *http.Request, database string, back string) {

	rows := getRows(r, database, "show tables")
	defer rows.Close()

	var n int = 1
	records := [][]string{}
	head := []string{"Table", "Rows"}

	for rows.Next() {
		var field string
		var row []string
		var nrows string
		rows.Scan(&field)
		if database == "information_schema" {
			nrows = "?"
		} else {
			nrows = getCount(r, database, field)
		}
		row = []string{href(r.URL.Host+"?"+r.URL.RawQuery+"&t="+field, strconv.Itoa(n)), field, nrows}
		records = append(records, row)
		n = n + 1
	}
	tableOut(w, r, back, head, records)
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
	
	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("action", "add")
	linkinsert := "/?" + q.Encode()
	q.Set("action", "subset")
	linkselect := "/?" + q.Encode()

	/*  credits:
	 * 	http://stackoverflow.com/questions/19991541/dumping-mysql-tables-to-json-with-golang
	 * 	http://go-database-sql.org/varcols.html
	 */
	raw := make([]interface{}, len(cols))
	val := make([]interface{}, len(cols))

	for i := range val {
		raw[i] = &val[i]
	}

	var n int = 1
	head := []string{}
	for _, column := range cols {
		head = append(head, column)
	}
	head = append(head, href(linkselect, "[?]") + href(linkinsert, "[+]"))
	
	records := [][]string{}
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
	tableOut(w, r, back, head, records)
}

// Dump all fields of a record, one column per line
func dumpFields(w http.ResponseWriter, r *http.Request, db string, t string, num string, back string) {

	rows := getRows(r, db, "select * from "+template.HTMLEscapeString(t))
	defer rows.Close()
	columns, err := rows.Columns()
	checkY(err)

	raw := make([]interface{}, len(columns))
	val := make([]interface{}, len(columns))

	for i := range val {
		raw[i] = &val[i]
	}

	rec, err := strconv.Atoi(num)
	checkY(err)
	nmax, err := strconv.Atoi(getCount(r, db, t))
	q := r.URL.Query()
	q.Set("n", strconv.Itoa(maxI(rec-1, 1)))
	linkleft := "?" + q.Encode()
	q.Set("n", strconv.Itoa(minI(rec+1, nmax)))
	linkright := "?" + q.Encode()

	q.Add("action", "add")
	linkinsert := "?" + q.Encode()
	q.Del("action")
	/*q.Add("action", "select")
	linkselect := "?" + q.Encode()
	q.Del("action")*/
	q.Add("action", "show")
	linkshow := "?" + q.Encode()
	q.Del("action")

	var n int = 1
	records := [][]string{}
	head := []string{"Field", "Content", href(linkleft, "[<]") + 
										 " [" + num + "] " + 
										 href(linkright, "[>]") +
										 " " +
										 href(linkinsert, "[+]") +
										 " " +
										 href(linkshow, "[?]")}

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
	tableOutFields(w, r, back, head, records)
}
