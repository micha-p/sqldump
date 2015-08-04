package main

import (
	"database/sql"
	"net/http"
	"net/url"
	"strings"
)


/* func showDatabases(w http.ResponseWriter, conn *sql.DB, host string) {
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
	tableOutSimple(w, conn, "", head, records, []Entry{})
}*/

func UpperAsColumn(array []sqlstring)sqlstring{
	var r sqlstring
	for _,v := range array{
		r=r + ", "+ str2sql(strings.ToUpper(sql2str(v)))+ " AS `"+v+"`"
	}
	return r
}

func showTableStatus(w http.ResponseWriter, conn *sql.DB, host string, db string,t string, o string, d string, g string, v string) {

	var query sqlstring
	query = "SELECT TABLE_NAME AS `Table`"
	
	if t=="1" {
		query = query + UpperAsColumn([]sqlstring{"Engine","Version","Row_format","Auto_increment"})
	} else if t=="2" {
		query = query + ", TABLE_ROWS AS `Rows`"
		query = query + UpperAsColumn([]sqlstring{"Avg_row_length","Data_length","Max_data_length","Index_length","Data_free"}) 
	} else if t=="3" {
		query = query + UpperAsColumn([]sqlstring{"Create_time","Update_time","Check_time"})
		query = query + ", TABLE_COLLATION AS `Collation`"
		query = query + UpperAsColumn([]sqlstring{"Checksum","Create_options"}) 
		query = query + ", TABLE_COMMENT AS `Comment`"
	} else {
		query = query + UpperAsColumn([]sqlstring{"Engine","Version","Row_format"})
		query = query + ", TABLE_ROWS AS `Rows`"
		query = query + UpperAsColumn([]sqlstring{"Avg_row_length","Data_length","Max_data_length","Index_length","Data_free"}) 
		query = query + UpperAsColumn([]sqlstring{"Auto_Increment"})
		query = query + UpperAsColumn([]sqlstring{"Create_time","Update_time","Check_time"})
		query = query + ", TABLE_COLLATION AS `Collation`"
		query = query + UpperAsColumn([]sqlstring{"Checksum","Create_options"}) 
		query = query + ", TABLE_COMMENT AS `Comment`"
	}
	query = query + " FROM information_schema.TABLES"
	
	m:=makeFreshQuery("", o, d)
	var menu []Entry							
	menu = append(menu, makeMenuPath(m, "t", "1", "1", "status"))
	menu = append(menu, makeMenuPath(m, "t", "2", "2", "status"))
	menu = append(menu, makeMenuPath(m, "t", "3", "3", "status"))
	menu = append(menu, makeMenuPath(m, "",  "", "i", "status"))

	showTableInfo(w, conn, host, db, t, o, d, g, v, "status", "SHOW TABLE STATUS",menu, query)
}

func showTables(w http.ResponseWriter, conn *sql.DB, host string, db string,t string, o string, d string, g string, v string) {
	query := str2sql("SELECT TABLE_NAME AS `Table`, TABLE_ROWS AS `Rows`, TABLE_COMMENT AS `Comment`")
	query = query + " FROM information_schema.TABLES"

	m:=makeFreshQuery("", o, d)
	var menu []Entry
	menu = append(menu, makeMenuPath(m, "", "", "i", "status"))

	showTableInfo(w, conn, host, db, t, o, d, g, v, "", "SHOW TABLES", menu, query)
}

func showTableInfo(w http.ResponseWriter, conn *sql.DB, host string, db string,
	t string, o string, d string, g string, v string, 
	path string, short sqlstring, menu []Entry, query sqlstring) {
	
	query = query + sqlWhere("TABLE_SCHEMA", "=", db) + sqlHaving(g, "=", v) + sqlOrder(o, d)
	rows, err, sec := getRows(conn, query)
	checkY(err)
	defer rows.Close()
	columns, err := rows.Columns()
	checkY(err)

	home := makeFreshQuery("", o, d)
	if g !=""{
		home.Add("g",g)
	}
	if v !=""{
		home.Add("v",v)
	}
	head := createHead("", o, d, "", "", columns, path, home)

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
		g.Add("t", getNullString(values[0]).String)
		row = append(row, escape(Int64toa(rownum), g.Encode()))
		g.Del("t")
		for i, c := range columns {
			nv := getNullString(values[i])
			if c == "Rows" && (INFOFLAG || EXPERTFLAG) && strings.ToUpper(db) == "INFORMATION_SCHEMA" {
				nv = sql.NullString{Valid: true, String: getCount(conn, sqlCount(row[1].Text))}
			}
			if c == "Table" || c == "Comment" { // these links should navigate, not group!
				v := nv.String
				g := url.Values{}
				g.Add("t", v)
				// these links should navigate, not group!
				row = append(row, escape(v, "", g.Encode()))
			} else {
				row = append(row, makeEntry(nv, c, "", path, home))
			}
		}
		records = append(records, row)
	}

	// Shortened statement
	query = short + sqlHaving(g, "=", v) + sqlOrder(o, d)
	var msg Message
	if QUIETFLAG {
		msg = Message{}
	} else {
		msg = Message{Msg: sql2str(query), Rows: rownum, Affected: -1, Seconds: sec}
	}
		
	tableOutRows(w, conn, "","","", "", "", "", Entry{}, Entry{}, head, records, menu, []Message{msg}, [][]Clause{})
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
func showColumns(w http.ResponseWriter, conn *sql.DB, t string, stmt sqlstring) {

	rows, err, _ := getRows(conn, stmt)
	checkY(err)
	defer rows.Close()

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
	// message supressed as it disturbs equal alignment of info, query and field.
	tableOutSimple(w, conn, t, head, records, []Entry{})
}

// do not export
// might modify query
func makeEntry(nv sql.NullString, column string, primary string, path string, q url.Values) Entry {
	if nv.Valid {
		v := nv.String
		if column == primary {
			q.Set("k", primary)
			q.Set("v", v)
			query := q.Encode()
			q.Del("k")
			q.Del("v")
			return escape(v, path, query)
		} else {
			q.Set("g", column)
			q.Set("v", v)
			query := q.Encode()
			q.Del("g")
			q.Del("v")
			return escape(v, path, query)
		}
	} else {
		return escapeNull()
	}
}
