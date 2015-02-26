package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
)

func getConnection(cred Access, database string) *sql.DB {
	conn, err := sql.Open(cred.Dbms, dsn(cred.User, cred.Pass, cred.Host, cred.Port, database))
	checkY(err)
	return conn
}
	

func getRows(cred Access, database string, stmt string) *sql.Rows {
	conn := getConnection(cred, database)
	defer conn.Close()

	statement, err := conn.Prepare(stmt)
	checkY(err)
	rows, err := statement.Query()
	checkY(err)

	return rows
}

func getCols(cred Access, database string, table string) []string {
	rows := getRows(cred, database, "select * from "+template.HTMLEscapeString(table))
	defer rows.Close()

	cols, err := rows.Columns()
	checkY(err)
	return cols
}

func getCount(cred Access, database string, table string) string {

	rows := getRows(cred, database, "select count(*) from "+template.HTMLEscapeString(table))
	defer rows.Close()

	rows.Next()
	var field string
	rows.Scan(&field)
	return field
}
