package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log"
)

func getConnection(cred Access, db string) *sql.DB {
	conn, err := sql.Open(cred.Dbms, dsn(cred.User, cred.Pass, cred.Host, cred.Port, db))
	checkY(err)
	return conn
}

func getRows(cred Access, db string, stmt string) *sql.Rows {
	conn := getConnection(cred, db)
	defer conn.Close()

	log.Println("SQL: " + stmt)
	statement, err := conn.Prepare(stmt)
	checkY(err)
	rows, err := statement.Query()
	checkY(err)

	return rows
}

func getCols(cred Access, db string, t string) []string {
	rows := getRows(cred, db, "select * from "+template.HTMLEscapeString(t))
	defer rows.Close()

	cols, err := rows.Columns()
	checkY(err)
	return cols
}

func getCount(cred Access, db string, t string) string {
	rows := getRows(cred, db, "select count(*) from "+template.HTMLEscapeString(t))
	defer rows.Close()

	rows.Next()
	var field string
	rows.Scan(&field)
	return field
}
