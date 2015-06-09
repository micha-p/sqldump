package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

// http://stackoverflow.com/questions/17845619/how-to-call-the-scan-variadic-function-in-golang-using-reflection/17885636#17885636
// http://blog.golang.org/laws-of-reflection

func getNullString(val interface{}) sql.NullString {
	b, ok := val.([]byte)
	if val == nil {
		return sql.NullString{"NULL", false}
	} else if ok {
		return sql.NullString{string(b), true}
	} else {
		return sql.NullString{fmt.Sprint(val), true}
	}
}

// This struct is filled with column info, but might also ship a single value
type CContext struct {
	Number    string
	Name      string
	Label     string
	IsNumeric string
	IsString  string
	Nullable  string
	Valid     string
	Value     string
	Readonly  string
}

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
	checkY(err)
	return getNullString(value).String
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
	log.Println("[SQL]", "get columns", db, t)
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

/*
func getMainType(t string) string {
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
		return "numeric"
	} else if cType || yType || bType || tType {
		return "string"
	} else {
		return ""
	}
}

func getColumnMainType(cred Access, db string, t string, c string) string {
	stmt := "select data_type from information_schema.columns where table_schema = '" + db + "'and table_name = '" + t + "' and column_name = '" + c+ "'"
	return getMainType(getSingleValue(cred, db, stmt))
}
*/

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

		nullable := ""
		if yes, _ := regexp.MatchString("YES", n); yes {
			nullable = "YES"
		}

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
			m = append(m, CContext{strconv.Itoa(i), f, f, "numeric", "", nullable, "", "", ""})
		} else if cType || yType || bType || tType {
			m = append(m, CContext{strconv.Itoa(i), f, f, "", "string", nullable, "", "", ""})
		} else {
			m = append(m, CContext{strconv.Itoa(i), f, f, "", "", nullable, "", "", ""})
		}
		i = i + 1
	}
	return m
}

func getNullStringMap(w http.ResponseWriter, db string, t string, cred Access, rows *sql.Rows) map[string]sql.NullString {

	vmap := make(map[string]sql.NullString)
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
		vmap[columns[i]] = getNullString(values[i])
	}
	return vmap
}
