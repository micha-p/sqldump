package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"net/http"
	"strconv"
)

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

	db := r.URL.Query().Get("db")
	t := r.URL.Query().Get("t")
	x := r.URL.Query().Get("x")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if db == "" {
		dumpHome(w, r)
	} else if t == "" {
		dumpTables(w, r, db)
	} else if x == "" {
		dumpRecords(w, r, db, t)
	} else {
		dumpFields(w, r, db, t, x)
	}
}

// Shows selection of databases at top level
func dumpHome(w http.ResponseWriter, r *http.Request) {

	rows := getRows(r, "", "show databases")
	defer rows.Close()

	var n int = 1
	records := [][]string{}
	for rows.Next() {
		var field string
		rows.Scan(&field)
		row := []string{href(r.URL.Host+"?"+"db="+field, "["+strconv.Itoa(n)+"]"), field}
		records = append(records, row)
		n = n + 1
	}

	tableOut(w, r, records)
}

//  Dump all tables of a database
func dumpTables(w http.ResponseWriter, r *http.Request, database string) {

	rows := getRows(r, database, "show tables")
	defer rows.Close()

	records := [][]string{[]string{"rows", "table"}}
	for rows.Next() {
		var field string
		var row []string
		rows.Scan(&field)
		if database == "information_schema" {
			row = []string{href(r.URL.Host+"?"+r.URL.RawQuery+"&t="+field, "?"), field}
		} else {
			row = []string{href(r.URL.Host+"?"+r.URL.RawQuery+"&t="+field, getCount(r, database, field)), field}
		}
		records = append(records, row)
	}
	tableOut(w, r, records)

}

//  Dump all records of a table, one per row
func dumpRecords(w http.ResponseWriter, r *http.Request, database string, table string) {

	rows := getRows(r, database, "select * from "+template.HTMLEscapeString(table))
	defer rows.Close()
	cols, err := rows.Columns()
	checkY(err)

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
	head := []string{"#"}
	for _, column := range cols {
		head = append(head, column)
	}
	records := [][]string{}
	records = append(records, head)
	for rows.Next() {

		row := []string{href(r.URL.Host+"?"+r.URL.RawQuery+"&x="+strconv.Itoa(n), strconv.Itoa(n))}

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
	tableOut(w, r, records)
}

// Dump all fields of a record, one column per line
func dumpFields(w http.ResponseWriter, r *http.Request, database string, table string, num string) {

	rows := getRows(r, database, "select * from "+template.HTMLEscapeString(table))
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

	var n int = 1
	records := [][]string{}

rowLoop:
	for rows.Next() {

		// unfortunately we have to iterate up to row of interest
		if n == rec {
			err = rows.Scan(raw...)
			checkY(err)

			for i, col := range val {
				var row []string
				if col != nil {
					row = []string{columns[i] + ":", string(col.([]byte))}
					records = append(records, row)
				}
			}
			break rowLoop
		}
		n = n + 1
	}
	tableOut(w, r, records)
}
