package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log"
	"regexp"
)

func getConnection(cred Access, db string) *sql.DB {
	conn, err := sql.Open(cred.Dbms, dsn(cred.User, cred.Pass, cred.Host, cred.Port, db))
	checkY(err)
	return conn
}

func getRows(cred Access, db string, stmt string) *sql.Rows {
	conn := getConnection(cred, db)
	defer conn.Close()

	log.Println("[SQL] " + stmt)
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

func getPrimary(cred Access, db string, t string) string {

	rows := getRows(cred, db, "show columns from "+template.HTMLEscapeString(t))
	defer rows.Close()

	primary := ""
	for rows.Next() {
		var f, t, n, k, e string
		var d []byte // or use http://golang.org/pkg/database/sql/#NullString
		err := rows.Scan(&f, &t, &n, &k, &d, &e)
		checkY(err)
		if k == "PRI" {
			primary = f
		}
	}
	return primary
}

func getNumericBool(cred Access, db string, t string, c string) bool {

	rows := getRows(cred, db, "show columns from "+template.HTMLEscapeString(t))
	defer rows.Close()

	for rows.Next() {
		var f, t, n, k, e string
		var d []byte // or use http://golang.org/pkg/database/sql/#NullString
		err := rows.Scan(&f, &t, &n, &k, &d, &e)
		checkY(err)
		if f == c {
			iType, _ := regexp.MatchString("int", t)
			fType, _ := regexp.MatchString("float", t)
			rType, _ := regexp.MatchString("real", t)
			dType, _ := regexp.MatchString("double", t)
			cType, _ := regexp.MatchString("decimal", t)
			nType, _ := regexp.MatchString("numeric", t)
			if iType || fType || rType || dType || cType || nType {
				return true
			} else {
				return false
			}
		}
	}
	log.Fatalln("column " + c + " not found")
	return false
}

func getSingle(cred Access, db string, q string) string {

	rows := getRows(cred, db, q)
	defer rows.Close()
	var value interface{}
	var valuePtr interface{}
	valuePtr = &value

rowLoop:
	for rows.Next() {
		// just one row
		err := rows.Scan(valuePtr)
		checkY(err)
		break rowLoop
	}
	return dumpValue(value)
}
