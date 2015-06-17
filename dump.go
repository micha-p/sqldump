package main

import (
	"database/sql"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

// TODO: better dispatching
func dumpSelection(w http.ResponseWriter, r *http.Request, conn *sql.DB,
	host string, db string, t string, o string, d string, n string, g string, k string, v string) {

	query := sqlStar(t)
	wclauses, _, whereQ := collectClauses(r, conn, t)

	if len(wclauses) > 0 {
		query = "SELECT TEMP.* FROM (" + query + sqlWhereClauses(wclauses) + ") TEMP "
	}
	if o != "" {
		query = query + sqlOrder(o, d)
	}

	if g !="" && v !=""{
		query = sqlStar(t) + sqlWhere(g, "=", v) + sqlOrder(o, d)
		dumpGroup(w, conn, host, db, t, o, d, g, v, query, whereQ)
	} else if n != "" {
		singlenumber := regexp.MustCompile("^ *(\\d+) *$").FindString(n)
		limits := regexp.MustCompile("^ *(\\d+) *- *(\\d+) *$").FindStringSubmatch(n)

		if singlenumber != "" {
			nint, _ := strconv.Atoi(singlenumber)
			query = query + sqlLimit(2, nint) // for finding next record
			dumpFields(w, conn, host, db, t, o, d, singlenumber, nint, query, whereQ)
		} else if len(limits) == 3 {
			startint, err := strconv.Atoi(limits[1])
			checkY(err)
			endint, err := strconv.Atoi(limits[2])
			checkY(err)
			maxint, err := strconv.Atoi(getCount(conn, t))
			checkY(err)
			endint = minI(endint, maxint)
			query = query + sqlLimit(1+endint-startint, startint)
			dumpRange(w, conn, host, db, t, o, d, startint, endint, maxint, query)
		} else {
			shipMessage(w, host, db, "Can't convert to number or range: "+n)
		}
	} else {
		if len(wclauses) > 0 {
			dumpWhere(w, conn, host, db, t, o, d, query, whereQ)
		} else {
			dumpRows(w, conn, host, db, t, o, d, query)
		}
	}
}

func dumpRows(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, query sqlstring) {

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)

	rows, err := getRows(conn, query)
	if err != nil {
		checkErrorPage(w, host, db, t, query, err)
		return
	} else {
		defer rows.Close()
	}

	primary := getPrimary(conn, t)
	columns, err := rows.Columns()
	checkY(err)
	head := createHead(db, t, o, d, "", primary, columns, q)

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
		q.Set("o", o)
		q.Set("d", d)
		q.Del("k")
		q.Del("v")
		q.Add("n", strconv.Itoa(rownum))
		row = append(row, escape(strconv.Itoa(rownum), q.Encode()))
		q.Del("n")

		err = rows.Scan(valuePtrs...)
		checkY(err)

		for i, _ := range columns {
			nv := getNullString(values[i])
			if nv.Valid {
				v := nv.String
				if columns[i] == primary {
					q.Del("o")
					q.Del("d")
					q.Del("n")
					q.Set("k", primary)
					q.Set("v", v)
					row = append(row, escape(v, q.Encode()))
					q.Del("k")
					q.Del("v")
				} else {
					g := url.Values{}
					g.Add("db", db)
					g.Add("t", t)
					g.Add("g", columns[i])
					g.Add("v", v)
					row = append(row, escape(v, g.Encode()))
				}
			} else {
				row = append(row, escapeNull())
			}
		}
		records = append(records, row)
	}

	q.Set("action", "QUERY")
	linkselect := q.Encode()
	q.Set("action", "ADD")
	linkinsert := q.Encode()
	q.Set("action", "UPDATEFORM")
	linkupdate := q.Encode()
	q.Set("action", "QUERYDELETE")
	linkdeleteF := q.Encode()
	q.Set("action", "INFO")
	linkinfo := q.Encode()
	q.Del("action")

	q.Del("n")
	linkleft := escape("<", q.Encode())
	linkright := escape(">", q.Encode())
	menu := []Entry{}
	menu = append(menu, escape("?", linkselect))
	menu = append(menu, escape("+", linkinsert))
	menu = append(menu, escape("~", linkupdate))
	menu = append(menu, escape("-", linkdeleteF))
	menu = append(menu, escape("i", linkinfo))

	var msg, nrows string
	if !QUIETFLAG {
		msg = sql2string(query)
		nrows = strconv.Itoa(rownum)
	}
	tableOutRows(w, conn, host, db, t, primary, o, d, " ", "#", linkleft, linkright, head, records, menu, msg, nrows, "", url.Values{})
}

func dumpGroup(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, g string, v string, query sqlstring, q url.Values) {

	q.Add("db", db)
	q.Add("t", t)
	q.Del("k")
	q.Del("v")

	rows, err := getRows(conn, query)
	if err != nil {
		checkErrorPage(w, host, db, t, query, err)
		return
	} else {
		defer rows.Close()
	}

	primary := getPrimary(conn, t)
	columns, err := rows.Columns()
	checkY(err)
	head := createHead(db, t, o, d, "", primary, columns, q)

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
		q.Set("o", o)
		q.Set("d", d)
		q.Add("n", strconv.Itoa(rownum))
		row = append(row, escape(strconv.Itoa(rownum), q.Encode()))
		q.Del("n")

		err = rows.Scan(valuePtrs...)
		checkY(err)

		for i, _ := range columns {
			nv := getNullString(values[i])
			if nv.Valid {
				v := nv.String
				if columns[i] == primary {
					q.Del("o")
					q.Del("d")
					q.Del("n")
					q.Set("k", primary)
					q.Set("v", v)
					q.Del("k")
					q.Del("v")
					row = append(row, escape(v, q.Encode()))
				} else {
					g := url.Values{}
					g.Add("db", db)
					g.Add("t", t)
					g.Add("g", columns[i])
					g.Add("v", v)
					row = append(row, escape(v, g.Encode()))
				}
			} else {
				row = append(row, escapeNull())
			}
		}

		records = append(records, row)
	}

	q.Set("action", "QUERY")
	linkselect := q.Encode()
	q.Set("action", "ADD")
	linkinsert := q.Encode()
/*	q.Set("action", "UPDATEFORM")
	linkupdate := q.Encode()
	q.Set("action", "DELETE")
	linkdelete := q.Encode() */
	q.Set("action", "INFO")
	linkinfo := q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, escape("?", linkselect))
	menu = append(menu, escape("+", linkinsert))
/*	menu = append(menu, escape("~", linkupdate))
	menu = append(menu, escape("-", linkdelete)) */
	menu = append(menu, escape("i", linkinfo))
	wherestring := WhereQuery2Pretty(q, getColumnInfo(conn, t))

	q.Set("g", g)
	next, err := getSingleValue(conn, host, db, sqlSelect(g, t)+sqlWhere(g, ">", v)+sqlOrder(g, "")+sqlLimit(1, 0))
	if err == nil {
		q.Set("v", next)
	} else {
		q.Set("v", v)
	}
	linkright := escape(">", q.Encode())
	prev, err := getSingleValue(conn, host, db, sqlSelect(g, t)+sqlWhere(g, "<", v)+sqlOrder(g, "1")+sqlLimit(1, 0))
	if err == nil {
		q.Set("v", prev)
	} else {
		q.Set("v", v)
	}
	linkleft := escape("<", q.Encode())

	var msg, nrows string
	if !QUIETFLAG {
		msg = sql2string(query)
		nrows = strconv.Itoa(rownum)
	}
	tableOutRows(w, conn, host, db, t, primary, o, d, v, g + " =" , linkleft, linkright, head, records, menu, msg, nrows, wherestring, q)
}

func dumpWhere(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, query sqlstring, q url.Values) {

	q.Add("db", db)
	q.Add("t", t)
	q.Del("k")
	q.Del("v")

	rows, err := getRows(conn, query)
	if err != nil {
		checkErrorPage(w, host, db, t, query, err)
		return
	} else {
		defer rows.Close()
	}

	primary := getPrimary(conn, t)
	columns, err := rows.Columns()
	checkY(err)
	head := createHead(db, t, o, d, "", primary, columns, q)

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
		q.Set("o", o)
		q.Set("d", d)
		q.Add("n", strconv.Itoa(rownum))
		row = append(row, escape(strconv.Itoa(rownum), q.Encode()))
		q.Del("n")

		err = rows.Scan(valuePtrs...)
		checkY(err)

		for i, _ := range columns {
			nv := getNullString(values[i])
			if nv.Valid {
				v := nv.String
				if columns[i] == primary {
					q.Del("o")
					q.Del("d")
					q.Del("n")
					q.Set("k", primary)
					q.Set("v", v)
					q.Del("k")
					q.Del("v")
					row = append(row, escape(v, q.Encode()))
				} else {
					g := url.Values{}
					g.Add("db", db)
					g.Add("t", t)
					g.Add("g", columns[i])
					g.Add("v", v)
					row = append(row, escape(v, g.Encode()))
				}
			} else {
				row = append(row, escapeNull())
			}
		}

		records = append(records, row)
	}

	q.Set("action", "QUERY")
	linkselect := q.Encode()
	q.Set("action", "ADD")
	linkinsert := q.Encode()
	q.Set("action", "UPDATEFORM")
	linkupdate := q.Encode()
	q.Set("action", "DELETE")
	linkdelete := q.Encode()
	q.Set("action", "INFO")
	linkinfo := q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, escape("?", linkselect))
	menu = append(menu, escape("+", linkinsert))
	menu = append(menu, escape("~", linkupdate))
	menu = append(menu, escape("-", linkdelete))
	menu = append(menu, escape("i", linkinfo))
	wherestring := WhereQuery2Pretty(q, getColumnInfo(conn, t))
	var msg, nrows string
	if !QUIETFLAG {
		msg = sql2string(query)
		nrows = strconv.Itoa(rownum)
	}
	tableOutRows(w, conn, host, db, t, primary, o, d, "", "", Entry{}, Entry{}, head, records, menu, msg, nrows, wherestring, q)
}

func dumpRange(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, start int, end int, max int, query sqlstring) {

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("action", "ADD")
	linkinsert := q.Encode()
	q.Set("action", "QUERY")
	linkselect := q.Encode()
	q.Set("action", "INFO")
	linkinfo := q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, escape("?", linkselect))
	menu = append(menu, escape("+", linkinsert))
	menu = append(menu, escape("i", linkinfo))

	limitstring := strconv.Itoa(start) + "-" + strconv.Itoa(end)
	q.Add("n", limitstring)

	rows, err := getRows(conn, query)
	if err != nil {
		checkErrorPage(w, host, db, t, query, err)
		return
	} else {
		defer rows.Close()
	}

	primary := getPrimary(conn, t)
	columns, err := rows.Columns()
	checkY(err)
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for i, _ := range columns {
		valuePtrs[i] = &values[i]
	}

	head := createHead(db, t, o, d, limitstring, "", columns, q)

	records := [][]Entry{}
	rowrange := end - start
	rownum := start -1
	for rows.Next() && rownum <= end {
		rownum = rownum + 1
		if o != "" {
			q.Set("o", o)
		}
		if d != "" {
			q.Set("d", d)
		}
		row := []Entry{}
		q.Set("n", strconv.Itoa(rownum))
		row = append(row, escape(strconv.Itoa(rownum), q.Encode()))

		err = rows.Scan(valuePtrs...)
		checkY(err)

		for i, _ := range columns {
			nv := getNullString(values[i])
			if nv.Valid {
				v := nv.String
				if columns[i] == primary {
					q.Del("o")
					q.Del("d")
					q.Del("n")
					q.Set("k", primary)
					q.Set("v", v)
					row = append(row, escape(v, q.Encode()))
				} else {
					g := url.Values{}
					g.Add("db", db)
					g.Add("t", t)
					g.Add("g", columns[i])
					g.Add("v", v)
					row = append(row, escape(v, g.Encode()))
				}
			} else {
				row = append(row, escapeNull())
			}
		}
		records = append(records, row)
	}

	q.Del("o")
	q.Del("d")
	q.Del("k")
	q.Del("v")
	left := maxI(start-rowrange, 1)
	right := minI(end+rowrange, max)
	q.Set("n", strconv.Itoa(left)+"-"+strconv.Itoa(left+rowrange-1))
	linkleft := escape("<", q.Encode())
	q.Set("n", strconv.Itoa(1+right-rowrange)+"-"+strconv.Itoa(right))
	linkright := escape(">", q.Encode())
	var msg, nrows string
	if !QUIETFLAG {
		msg = sql2string(query)
		nrows = strconv.Itoa(rownum)
	}
	tableOutRows(w, conn, host, db, t, primary, o, d, limitstring, "#", linkleft, linkright, head, records, menu, msg, nrows, "", url.Values{})
}
