package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

// retrieving column information might be combined better

func getConnection(cred Access, db string) *sql.DB {
	conn, err := sql.Open(cred.Dbms, dsn(cred.User, cred.Pass, cred.Host, cred.Port, db))
	checkY(err)
	return conn
}

func getRows(cred Access, db string, stmt string) (*sql.Rows, error) {
	conn := getConnection(cred, db)
	defer conn.Close()

	log.Println("[SQL]", stmt)
	rows, err := conn.Query(stmt)
	return rows, err
}

func getSingleValue(cred Access, db string, stmt string) string {
	conn := getConnection(cred, db)
	defer conn.Close()
	log.Println("[SQL]", stmt)
	row := conn.QueryRow(stmt)

	var value interface{}
	var valuePtr interface{}
	valuePtr = &value
	err := row.Scan(valuePtr)

	if err == nil {
		return dumpValue(value)
	} else {
		return "NULL"
	}
}

func getCount(cred Access, db string, t string) string {

	countstmt := "select count(*) from `" + t + "`"
	conn := getConnection(cred, db)
	defer conn.Close()
	log.Println("[SQL]", countstmt)
	// rows,err := conn.Query("select count(*) from ?", t) // does not work??
	row := conn.QueryRow(countstmt)

	var field string
	row.Scan(&field)
	return field
}

func getCols(cred Access, db string, t string) []string {

	conn := getConnection(cred, db)
	defer conn.Close()
	log.Println("[SQL]", "get columns",db,t)
	// rows, err := conn.Query("select * from ? limit 1") // does not work??
	rows, err := conn.Query("select * from `" + t + "` limit 0")
	checkY(err)
	defer rows.Close()

	cols, err := rows.Columns()
	return cols
}

func getPrimary(cred Access, db string, t string) string {

	conn := getConnection(cred, db)
	defer conn.Close()
	// rows, err := conn.Query("show columns from ?", t) // does not work??
	rows, err := conn.Query("show columns from `" + t + "`")
	checkY(err)
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

func getColumnInfo(cred Access, db string, t string) []CContext {

	conn := getConnection(cred, db)
	defer conn.Close()
	rows, err := conn.Query("show columns from `" + t + "`")
	checkY(err)
	defer rows.Close()

	m := []CContext{}
	i := 1
	for rows.Next() {
		var f, t, n, k, e string
		var d []byte // or use http://golang.org/pkg/database/sql/#NullString
		err := rows.Scan(&f, &t, &n, &k, &d, &e)
		checkY(err)

		iType, _ := regexp.MatchString("int", t)
		fType, _ := regexp.MatchString("float", t)
		rType, _ := regexp.MatchString("real", t)
		dType, _ := regexp.MatchString("double", t)
		lType, _ := regexp.MatchString("decimal", t)
		nType, _ := regexp.MatchString("numeric", t)
		cType, _ := regexp.MatchString("char", t)
		yType, _ := regexp.MatchString("binary", t)
		bType, _ := regexp.MatchString("blob", t)
		tType, _ := regexp.MatchString("text", t)

		if iType || fType || rType || dType || lType || nType {
			m = append(m, CContext{strconv.Itoa(i), f, f, "numeric", "", "", ""})
		} else if cType || yType || bType || tType {
			m = append(m, CContext{strconv.Itoa(i), f, f, "", "string", "", ""})
		} else {
			m = append(m, CContext{strconv.Itoa(i), f, f, "", "", "", ""})
		}
		i = i + 1
	}
	return m
}

func getFieldMap(w http.ResponseWriter, db string, t string, cred Access, rows *sql.Rows) map[string]string {

	fieldmap := make(map[string]string)
	defer rows.Close()

	columns, err := rows.Columns()
	checkY(err)
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for i, _ := range columns {
		valuePtrs[i] = &values[i]
	}

	rows.Next() // just one row
	err = rows.Scan(valuePtrs...)
	checkY(err)

	for i, _ := range columns {
		fieldmap[columns[i]] = dumpValue(values[i])
	}
	return fieldmap
}
