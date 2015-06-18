package main

import (
	"database/sql"
	"net/http"
	"net/url"
	"regexp"
)

// rows are not adressable:
// dumpRows  -> SELECTFORM, INSERTFORM, INFO
// dumpRange -> SELECTFORM, INSERTFORM, INFO
// dumpField -> SELECTFORM, INSERTFORM, INFO

// rows are selected by where-clause
// dumpWhere 		-> SELECTFORM, INSERTFORM, UPDATEFORM, DELETE, INFO

// rows are selected by key or group
// dumpKeyValue 	-> SELECTFORM, INSERTFORM, UPDATEFORM, DELETE, INFO
// dumpGroup	 	-> SELECTFORM, INSERTFORM, UPDATEFORM, DELETE, INFO


func makeEntry(nv sql.NullString, db string, t string, c string, primary string) Entry {
	if nv.Valid {
		v := nv.String
		g := url.Values{}
		g.Add("db", db)
		g.Add("t", t)
		if c == primary {
			g.Add("k", primary)
			g.Add("v", v)
			return escape(v, g.Encode())
		} else {
			g.Add("g", c)
			g.Add("v", v)
			return escape(v, g.Encode())
		}
	} else {
		return escapeNull()
	}
}


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
		query = sqlStar(t) + sqlWhereClauses(wclauses) + sqlHaving(g, "=", v) + sqlOrder(o, d)
		dumpGroup(w, conn, host, db, t, o, d, g, v, query, whereQ)
	} else if n != "" {
		singlenumber := regexp.MustCompile("^ *(\\d+) *$").FindString(n)
		limits := regexp.MustCompile("^ *(\\d+) *- *(\\d+) *$").FindStringSubmatch(n)

		if singlenumber != "" {
			nint, _ := Atoi64(singlenumber)
			query = query + sqlLimit(2, nint) // for finding next record
			dumpFields(w, conn, host, db, t, o, d, singlenumber, nint, query, whereQ)
		} else if len(limits) == 3 {
			startint, err := Atoi64(limits[1])
			checkY(err)
			endint, err := Atoi64(limits[2])
			checkY(err)
			maxint, err := Atoi64(getCount(conn, t))
			checkY(err)
			endint = minInt64(endint, maxint)
			query = query + sqlLimit(1+endint-startint, startint)
			dumpRange(w, conn, host, db, t, o, d, startint, endint, maxint, query)
		} else {
			shipMessage(w, host, db, "Can't convert to number or range: "+n)
		}
	} else {
		if len(wclauses) > 0 {
			dumpWhere(w, conn, host, db, t, o, d, query, whereQ)
		} else {
			dumpRows(w, conn, host, db, t, o, d, []Message{}, query)
		}
	}
}


func dumpRows(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, messageStack []Message, query sqlstring) {

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)

	rows, err, sec := getRows(conn, query)
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
	var rownum int64 = 0
	for rows.Next() {
		rownum = rownum + 1
		row := []Entry{}
		q.Set("o", o)
		q.Set("d", d)
		q.Del("k")
		q.Del("v")
		q.Add("n", Int64toa(rownum))
		row = append(row, escape(Int64toa(rownum), q.Encode()))
		q.Del("n")

		err = rows.Scan(valuePtrs...)
		checkY(err)

		for i, c := range columns {
			nv := getNullString(values[i])
			row = append(row, makeEntry(nv, db, t, c, primary))
		}
		records = append(records, row)
	}

	q.Set("action", "SELECTFORM")
	linkselect := q.Encode()
	q.Set("action", "INSERTFORM")
	linkinsert := q.Encode()
	q.Set("action", "SELECTFORM")
	linkupdate := q.Encode()
	q.Set("action", "DELETEFORM")
	linkdelete := q.Encode()
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
	menu = append(menu, escape("-", linkdelete))
	menu = append(menu, escape("i", linkinfo))

	var msg Message
	if QUIETFLAG {
		msg = Message{}
	} else {
		msg = Message{Msg:sql2string(query),Rows:rownum,Affected:-1,Seconds:sec }
	}
	messageStack = append(messageStack,msg)
	tableOutRows(w, conn, host, db, t, primary, o, d, " ", "#", linkleft, linkright, head, records, menu, messageStack, "", url.Values{})
}

// difference to dumprows
// 1. counter, label, linkleft and linkright
// 2. as there is already a selection, update will show UPDATEFORM
// 3. Delete will delete immediately

func dumpGroup(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, g string, v string, query sqlstring, q url.Values) {

	q.Add("db", db)
	q.Add("t", t)
	q.Add("g", g)
	q.Add("v", v)
	q.Del("k")

	rows, err, sec := getRows(conn, query)
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
	var rownum int64 =  0
	for rows.Next() {
		rownum = rownum + 1
		row := []Entry{}
		q.Set("o", o)
		q.Set("d", d)
		q.Add("n", Int64toa(rownum))
		row = append(row, escape(Int64toa(rownum), q.Encode()))
		q.Del("n")

		err = rows.Scan(valuePtrs...)
		checkY(err)

		for i, c := range columns {
			nv := getNullString(values[i])
			row = append(row, makeEntry(nv, db, t, c, primary))
		}
		records = append(records, row)
	}

	q.Add("g", g)
	q.Set("action", "SELECTFORM")
	linkselect := q.Encode()
	q.Set("action", "INSERTFORM")
	linkinsert := q.Encode()
	q.Set("action", "GV_UPDATEFORM")
	linkupdate := q.Encode()
	q.Set("action", "GV_DELETE")
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

	var msg Message
	if QUIETFLAG {
		msg = Message{}
	} else {
		msg = Message{Msg:sql2string(query),Rows:rownum,Affected:-1,Seconds:sec }
	}
	tableOutRows(w, conn, host, db, t, primary, o, d, v, g + " =" , linkleft, linkright, head, records, menu, []Message{msg},wherestring, q)
}

// difference to dumprows
// 1. trail shows where clauses
// 2. as there is already a selection, update will show UPDATEFORM
// 3. delete will show FILLEDDELETEFORM for confirmation (TODO)

func dumpWhere(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, query sqlstring, q url.Values) {

	q.Add("db", db)
	q.Add("t", t)
	q.Del("k")
	q.Del("v")

	rows, err, sec := getRows(conn, query)
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
	var rownum int64 = 0
	for rows.Next() {
		rownum = rownum + 1
		row := []Entry{}
		q.Set("o", o)
		q.Set("d", d)
		q.Add("n", Int64toa(rownum))
		row = append(row, escape(Int64toa(rownum), q.Encode()))
		q.Del("n")

		err = rows.Scan(valuePtrs...)
		checkY(err)

		for i, c := range columns {
			nv := getNullString(values[i])
			row = append(row, makeEntry(nv, db, t, c, primary))
		}

		records = append(records, row)
	}

	q.Set("action", "SELECTFORM")
	linkselect := q.Encode()
	q.Set("action", "INSERTFORM")
	linkinsert := q.Encode()
	q.Set("action", "UPDATEFORM")
	linkupdate := q.Encode()
	q.Set("action", "FILLEDDELETEFORM")
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
	var msg Message
	if QUIETFLAG {
		msg = Message{}
	} else {
		msg = Message{Msg:sql2string(query),Rows:rownum,Affected:-1,Seconds:sec }
	}
	tableOutRows(w, conn, host, db, t, primary, o, d, "", "", Entry{}, Entry{}, head, records, menu, []Message{msg}, wherestring, q)
}

// as this is not a selection based on where clauses, manipulation is not possible
func dumpRange(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, start int64, end int64, max int64, query sqlstring) {

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("action", "INSERTFORM")
	linkinsert := q.Encode()
	q.Set("action", "SELECTFORM")
	linkselect := q.Encode()
	q.Set("action", "INFO")
	linkinfo := q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, escape("?", linkselect))
	menu = append(menu, escape("+", linkinsert))
	menu = append(menu, escape("i", linkinfo))

	limitstring := Int64toa(start) + "-" + Int64toa(end)
	q.Add("n", limitstring)

	rows, err, sec := getRows(conn, query)
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
		q.Set("n", Int64toa(rownum))
		row = append(row, escape(Int64toa(rownum), q.Encode()))

		err = rows.Scan(valuePtrs...)
		checkY(err)

		for i, c := range columns {
			nv := getNullString(values[i])
			row = append(row, makeEntry(nv, db, t, c, primary))
		}
		records = append(records, row)
	}

	q.Del("o")
	q.Del("d")
	q.Del("k")
	q.Del("v")
	left := maxInt64(start-rowrange, 1)
	right := minInt64(end+rowrange, max)
	q.Set("n", Int64toa(left)+"-"+Int64toa(left+rowrange-1))
	linkleft := escape("<", q.Encode())
	q.Set("n", Int64toa(1+right-rowrange)+"-"+Int64toa(right))
	linkright := escape(">", q.Encode())

	var msg Message
	if QUIETFLAG {
		msg = Message{}
	} else {
		msg = Message{Msg:sql2string(query),Rows:rownum,Affected:-1,Seconds:sec }
	}
	tableOutRows(w, conn, host, db, t, primary, o, d, limitstring, "#", linkleft, linkright, head, records, menu, []Message{msg}, "", url.Values{})
}
