package main

import (
	"database/sql"
	"net/http"
	"net/url"
)

func showDatabases(w http.ResponseWriter, conn *sql.DB, host string) {

	q := url.Values{}
	// "SELECT TABLE_NAME AS `Table`, ENGINE AS `Engine`, TABLE_ROWS AS `Rows`,TABLE_COLLATION AS `Collation`,CREATE_TIME AS `Create`, TABLE_COMMENT AS `Comment`
	stmt := str2sql("SHOW DATABASES")
	rows, err, _ := getRows(conn, stmt)
	checkY(err)
	defer rows.Close()

	records := [][]Entry{}
	head := []Entry{{"#", "", ""}, {"Database", "", ""}}
	var n int64
	for rows.Next() {
		n = n + 1
		var field string
		rows.Scan(&field)
		if EXPERTFLAG || INFOFLAG || field != "information_schema" {
			q.Set("db", field)
			link := q.Encode()
			row := []Entry{escape(Int64toa(n), link), escape(field, link)}
			records = append(records, row)
		}
	}
	// message suppressed, as it is not really useful and database should be chosen at login or bookmarked
	tableOutSimple(w, conn, host, "", "", head, records, []Entry{})
}

func showTables(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, g string, v string) {

	q := url.Values{}
	q.Add("db", db)
	query := str2sql("SELECT TABLE_NAME AS `Table`, TABLE_ROWS AS `Rows`, TABLE_COMMENT AS `Comment`")
	query = query + " FROM information_schema.TABLES"
	query = query + sqlWhere("TABLE_SCHEMA", "=", db) + sqlHaving(g, "=", v) + sqlOrder(o, d)
	rows, err, sec := getRows(conn, query)
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
	var rownum int64
	for rows.Next() {
		rownum = rownum + 1
		row := []Entry{}
		err = rows.Scan(valuePtrs...)
		checkY(err)
		g := url.Values{}
		g.Add("db", db)
		g.Add("t", getNullString(values[0]).String)
		row = append(row, escape(Int64toa(rownum), g.Encode()))
		for i, c := range columns {
			nv := getNullString(values[i])
			if c == "Rows" && (db == "INFORMATION_SCHEMA" || db == "information_schema") && (INFOFLAG || EXPERTFLAG) {
				nv = sql.NullString{Valid: true, String: getCount(conn, sqlCount(row[1].Text))}
			}
			if c == "Table" || c == "Comment" {
				v := nv.String
				g := url.Values{}
				g.Add("db", db)
				g.Add("t", v)
				row = append(row, escape(v, g.Encode()))
			} else {
				row = append(row, makeEntry(nv, db, "", c, "",q))
			}
		}
		records = append(records, row)
	}

	// Shortened statement
	query = "SHOW TABLES" + sqlHaving(g, "=", v) + sqlOrder(o, d)
	var msg Message
	if QUIETFLAG {
		msg = Message{}
	} else {
		msg = Message{Msg: sql2str(query), Rows: rownum, Affected: -1, Seconds: sec}
	}
	tableOutRows(w, conn, host, db, "", "", "", "", "", "", Entry{}, Entry{}, head, records, []Entry{}, []Message{msg}, "", url.Values{})
}

/*
 show columns from posts;
+-------+-------------+------+-----+---------+-------+
| Field | Type        | Null | Key | Default | Extra |
+-------+-------------+------+-----+---------+-------+
| title | varchar(64) | YES  |     | NULL    |       |
| start | date        | YES  |     | NULL    |       |
+-------+-------------+------+-----+---------+-------+
*/
func showInfo(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, stmt sqlstring) {

	rows, err, _ := getRows(conn, stmt)
	checkY(err)
	defer rows.Close()

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Set("action", "SELECTFORM")
	linkselect := q.Encode()
	q.Set("action", "INSERTFORM")
	linkinsert := q.Encode()
	q.Set("action", "DELETEFORM")
	linkdeleteF := q.Encode()
	q.Set("action", "INFO")
	linkinfo := q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, escape("?", linkselect))
	menu = append(menu, escape("+", linkinsert))
	menu = append(menu, escape("-", linkdeleteF))
	menu = append(menu, escape("i", linkinfo))

	records := [][]Entry{}
	head := []Entry{escape("#"), escape("Field"), escape("Type"), escape("Null"), escape("Key"), escape("Default"), escape("Extra")}

	var i int64 = 1
	for rows.Next() {
		var f, t, n, k, e string
		var d []byte // or use http://golang.org/pkg/database/sql/#NullString
		err := rows.Scan(&f, &t, &n, &k, &d, &e)
		checkY(err)
		records = append(records, []Entry{escape(Int64toa(i)), escape(f), escape(t), escape(n), escape(k), escape(string(d)), escape(e)})
		i = i + 1
	}
	// message not shown as it disturbs equal alignment of info, query and field.
	tableOutSimple(w, conn, host, db, t, head, records, menu)
}

// do not export
// will modify query
func makeEntry(nv sql.NullString, db string, t string, c string, primary string, q url.Values) Entry {
	if nv.Valid {
		v := nv.String
		q.Set("db", db)
		q.Set("t", t)
		if c == primary {
			q.Set("k", primary)
			q.Set("v", v)
			return escape(v, q.Encode())
		} else {
			q.Set("g", c)
			q.Set("v", v)
			return escape(v, q.Encode())
		}
	} else {
		return escapeNull()
	}
}
