package main

import (
	"database/sql"
	"net/http"
	"net/url"
	"strconv"
)

func dumpIt(w http.ResponseWriter, r *http.Request, conn *sql.DB,
	host string, db string, t string, o string, d string, n string, g string, k string, v string) {

	if db == "" {
		dumpHome(w, conn, host)
		return
	} else if t == "" {
		dumpTables(w, conn, host, db, t, o, d, g, v)
	} else if k != "" && v != "" && k == getPrimary(conn, t) {
		dumpKeyValue(w, db, t, k, v, conn, host, sqlStar(t)+sqlWhere(k, "=", v))
	} else {
		dumpSelection(w, r, conn, host, db, t, o, d, n, g, k, v)
	}
}

// Shows selection of databases at top level
func dumpHome(w http.ResponseWriter, conn *sql.DB, host string) {

	q := url.Values{}
	stmt := string2sql("SHOW DATABASES")
	rows, err, _ := getRows(conn, stmt)
	checkY(err)
	defer rows.Close()

	records := [][]Entry{}
	head := []Entry{{"#", "", ""}, {"Database", "", ""}}
	var n int = 1
	for rows.Next() {
		var field string
		rows.Scan(&field)
		if EXPERTFLAG || INFOFLAG || field != "information_schema" {
			q.Set("db", field)
			link := q.Encode()
			row := []Entry{escape(strconv.Itoa(n), link), escape(field, link)}
			records = append(records, row)
			n = n + 1
		}
	}
	// message suppressed, as it is not really useful and database should be chosen at login or bookmarked
	tableOutSimple(w, conn, host, "", "", head, records, []Entry{})
}

//  Dump all tables of a database
func dumpTables(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, g string, v string) {

	q := url.Values{}
	q.Add("db", db)
    // "SELECT TABLE_NAME AS `Table`, ENGINE AS `Engine`, TABLE_ROWS AS `Rows`,TABLE_COLLATION AS `Collation`,CREATE_TIME AS `Create`, TABLE_COMMENT AS `Comment`
	stmt := string2sql("SELECT TABLE_NAME AS `Table`, TABLE_ROWS AS `Rows`, TABLE_COMMENT AS `Comment`")

	stmt = stmt + " FROM information_schema.TABLES"
	stmt = stmt + sqlWhere("TABLE_SCHEMA","=",db) + sqlHaving(g, "=", v) + sqlOrder(o,d)
	rows, err, sec := getRows(conn, stmt)
	checkY(err)
	defer rows.Close()

	columns, err := rows.Columns()
	checkY(err)
	home := url.Values{}
	home.Add("db", db)
	home.Add("o", o)
	home.Add("d", d)
	head := createHead(db, "", o, d, "", "", columns, home)

	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for i, _ := range columns {
		valuePtrs[i] = &values[i]
	}
	records := [][]Entry{}
	rownum := 0
	for rows.Next() {
		rownum = rownum + 1
		row := []Entry{}
		err = rows.Scan(valuePtrs...)
		checkY(err)
		g := url.Values{}
		g.Add("db", db)
		g.Add("t", getNullString(values[0]).String)
		row = append(row, escape(strconv.Itoa(rownum), g.Encode()))
		for i, c := range columns {
			nv := getNullString(values[i])
			if c == "Table" || c == "Comment" {
				v := nv.String
				g := url.Values{}
				g.Add("db", db)
				g.Add("t", v)
				row = append(row, escape(v, g.Encode()))
			} else if c == "Rows" && (db == "INFORMATION_SCHEMA" || db =="information_schema") && (INFOFLAG || EXPERTFLAG) {
				v := getCount(conn,row[1].Text)
				g := url.Values{}
				g.Add("db", db)
				g.Add("g", c)
				g.Add("v", v)
				row = append(row, escape(v, g.Encode()))
			} else if nv.Valid {
				v := nv.String
				g := url.Values{}
				g.Add("db", db)
				g.Add("g", c)
				g.Add("v", v)
				row = append(row, escape(v, g.Encode()))
			} else {
				row = append(row, escapeNull())
			}
		}
		records = append(records, row)
	}

	// Shortened statement
	stmt = "SHOW TABLES" + sqlHaving(g, "=", v) + sqlOrder(o,d)
	tableOutRows(w, conn, host, db, "", "", "", "", "", "", Entry{}, Entry{}, head, records, []Entry{}, sql2string(stmt), rownum, -1, sec,"", url.Values{})
}
