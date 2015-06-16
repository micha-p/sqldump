package main

import (
	"database/sql"
	"fmt"
	"html"
	"log"
	"regexp"
	"strconv"
)

// http://stackoverflow.com/questions/17845619/how-to-call-the-scan-variadic-function-in-golang-using-reflection/17885636#17885636
// http://blog.golang.org/laws-of-reflection

func getNullString(val interface{}) sql.NullString {
	b, ok := val.([]byte)
	if val == nil {
		return sql.NullString{"", false}
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

// TODO remove host, db as conn is db specific
func getRows(conn *sql.DB, stmt sqlstring) (*sql.Rows, error) {
	err := conn.Ping()
	checkY(err)
	log.Println("[SQL]", sql2string(stmt))
	rows, err := sqlQuery(conn, stmt)
	return rows, err
}

func getSingleValue(conn *sql.DB, host string, db string, stmt sqlstring) (string, error) {
	log.Println("[SQL]", sql2string(stmt))
	err := conn.Ping()
	checkY(err)
	row := sqlQueryRow(conn, stmt)

	var value interface{}
	var valuePtr interface{}
	valuePtr = &value
	err = row.Scan(valuePtr)
	return getNullString(value).String, err
}

func getCount(conn *sql.DB, t string) string {
	stmt := sqlCount(t)
	log.Println("[SQL]", sql2string(stmt))
	err := conn.Ping()
	checkY(err)
	row := sqlQueryRow(conn, stmt)

	var field string
	row.Scan(&field)
	return field
}

// retrieves column names from empty set
// only needed for collecting clauses
// TODO change parameters for clauses

func getCols(conn *sql.DB, t string) []string {

	log.Println("[SQL]", "get columns quickly: ", t)
	err := conn.Ping()
	checkY(err)
	rows, err := sqlQuery(conn, sqlStar(t)+sqlLimit(0, 0))
	checkY(err)
	defer rows.Close()

	cols, err := rows.Columns()
	return cols
}

func getPrimary(conn *sql.DB, t string) string {
	err := conn.Ping()
	checkY(err)
	rows, err := sqlQuery(conn, sqlColumns(t)+sqlWhere("Key", "=", "PRI"))
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

func getColumnMainType(conn *sql.DB, host string, db string, t string, c string) string {
	stmt := "select data_type from information_schema.columns where table_schema = '" + db + "'and table_name = '" + t + "' and column_name = '" + c+ "'"
	return getMainType(getSingleValue(conn, host, db, stmt))
}
*/

func getColumnInfo(conn *sql.DB, t string) []CContext {
	err := conn.Ping()
	checkY(err)
	rows, err := sqlQuery(conn, sqlColumns(t))
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

func getColumnInfoFilled(conn *sql.DB, host string, db string, t string, primary string, rows *sql.Rows) []CContext {

	err := conn.Ping()
	checkY(err)

	// TODO more efficient
	cols := getColumnInfo(conn, t)
	vmap := getNullStringMap(rows)

	newcols := []CContext{}
	for _, col := range cols {
		name := html.EscapeString(col.Name)
		readonly := ""
		value := html.EscapeString(vmap[col.Name].String)
		valid := ""
		if len(vmap) == 0 || vmap[col.Name].Valid {
			valid = "valid"
		}
		if name == primary {
			readonly = "1"
		}
		newcols = append(newcols, CContext{col.Number, name, name, col.IsNumeric, col.IsString, col.Nullable, valid, value, readonly})
	}
	return newcols
}

func getNullStringMap(rows *sql.Rows) map[string]sql.NullString {

	vmap := make(map[string]sql.NullString)
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
