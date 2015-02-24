package main

import (
	"html/template"
	"net/http"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
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
