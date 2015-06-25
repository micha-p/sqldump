package main

import (
	"database/sql"
	"net/http"
	"net/url"
)

/* Outline of routines:
 *
 * get pointer to rows
 * get further data
 * create head
 * init values and valuePtrs
 * build records
 * create menu
 * push message
 * table out
 */

func dumpRows(w http.ResponseWriter, conn *sql.DB, t string, o string, d string, stmt sqlstring, messageStack []Message) {

	menu := makeMenu3(makeFreshQuery(t, o, d))
	q := makeFreshQuery(t, o, d)
	linkleft := escape("<", q.Encode())
	linkright := escape(">", q.Encode())
	rows, err, sec := getRows(conn, stmt)
	if err != nil {
		checkErrorPage(w, conn, t, stmt, err)
		return
	} else {
		defer rows.Close()
	}

	primary := getPrimary(conn, t)
	columns, err := rows.Columns()
	checkY(err)
	head := createHead(t, o, d, "", primary, columns, url.Values{})
	records, rownum := makeRecords(rows, t, primary, 0, q)

	messageStack = append(messageStack, Message{Msg: sql2str(stmt), Rows: rownum, Affected: -1, Seconds: sec})
	tableOutRows(w, conn, t, primary, o, d, " ", "#", linkleft, linkright, head, records, menu, messageStack, [][]Clause{})
}

// difference to dumprows
// 1. counter, label, linkleft and linkright
// 2. as there is already a selection, update will show UPDATEFORM
// 3. Delete will delete immediately
func dumpGroup(w http.ResponseWriter, conn *sql.DB, t string, o string, d string, g string, v string,
	stmt sqlstring, whereStack [][]Clause, messageStack []Message) {

	q := makeFreshQuery(t, o, d)
	q.Set("g", g)
	q.Set("v", v)
	putWhereStackIntoQuery(q, whereStack)
	menu := makeMenu5(q)
	rows, err, sec := getRows(conn, stmt)
	if err != nil {
		checkErrorPage(w, conn, t, stmt, err)
		return
	} else {
		defer rows.Close()
	}

	/********** do this first to ensure correct query */
	var linkleft, linkright Entry
	{
		next, err := getSingleValue(conn, sqlSelect(g, t)+sqlWhereClauses(whereStack)+sqlHaving(g, ">", v)+sqlLimit(1, 0))
		if err == nil {
			q.Set("v", next)
		} else {
			q.Set("v", v)
		}
		linkright = escape(">", q.Encode())
		prev, err := getSingleValue(conn, sqlSelect(g, t)+sqlWhereClauses(whereStack)+sqlHaving(g, "<", v)+sqlOrder(g, "1")+sqlLimit(1, 0))
		if err == nil {
			q.Set("v", prev)
		} else {
			q.Set("v", v)
		}
		linkleft = escape("<", q.Encode())
		q.Set("v", v)
	}

	primary := getPrimary(conn, t)
	columns, err := rows.Columns()
	checkY(err)
	head := createHead(t, o, d, "", primary, columns, q)
	records, rownum := makeRecords(rows, t, primary, 0, q)

	messageStack = append(messageStack, Message{Msg: sql2str(stmt), Rows: rownum, Affected: -1, Seconds: sec})
	tableOutRows(w, conn, t, primary, o, d, v, g+" =", linkleft, linkright, head, records, menu, messageStack, whereStack)
}

// difference to dumpRows
// 1. trail shows where clauses
// 2. as there is already a selection, update will show UPDATEFORM
// 3. delete will show FILLEDDELETEFORM for confirmation (TODO)
func dumpWhere(w http.ResponseWriter, conn *sql.DB, t string, o string, d string,
	stmt sqlstring, whereStack [][]Clause, messageStack []Message) {

	q := makeFreshQuery(t, o, d)
	putWhereStackIntoQuery(q, whereStack)
	menu := makeMenu5(q)
	rows, err, sec := getRows(conn, stmt)
	if err != nil {
		checkErrorPage(w, conn, t, stmt, err)
		return
	} else {
		defer rows.Close()
	}

	primary := getPrimary(conn, t)
	columns, err := rows.Columns()
	checkY(err)
	head := createHead(t, o, d, "", primary, columns, q)
	records, rownum := makeRecords(rows, t, primary, 0, q)

	messageStack = append(messageStack, Message{Msg: sql2str(stmt), Rows: rownum, Affected: -1, Seconds: sec})
	tableOutRows(w, conn, t, primary, o, d, "", "", Entry{}, Entry{}, head, records, menu, messageStack, whereStack)
}

func dumpRange(w http.ResponseWriter, conn *sql.DB, t string, o string, d string, start int64, end int64, max int64,
	stmt sqlstring, whereStack [][]Clause, messageStack []Message) {

	q := makeFreshQuery(t, o, d)
	putWhereStackIntoQuery(q, whereStack)

	rowrange := end - start
	left := maxInt64(start-rowrange, 1)
	right := minInt64(end+rowrange, max)
	q.Set("n", Int64toa(left)+"-"+Int64toa(left+rowrange-1))
	linkleft := escape("<", q.Encode())
	q.Set("n", Int64toa(1+right-rowrange)+"-"+Int64toa(right))
	linkright := escape(">", q.Encode())

	limitstring := Int64toa(start) + "-" + Int64toa(end)
	q.Set("n", limitstring)
	menu := makeMenu3(q)

	rows, err, sec := getRows(conn, stmt)
	if err != nil {
		checkErrorPage(w, conn, t, stmt, err)
		return
	} else {
		defer rows.Close()
	}

	primary := getPrimary(conn, t)
	columns, err := rows.Columns()
	checkY(err)
	head := createHead(t, o, d, limitstring, "", columns, q)
	records, rownum := makeRecords(rows, t, primary, start-1, q)

	messageStack = append(messageStack, Message{Msg: sql2str(stmt), Rows: rownum, Affected: -1, Seconds: sec})
	tableOutRows(w, conn, t, primary, o, d, limitstring, "#", linkleft, linkright, head, records, menu, messageStack, whereStack)
}

/**** HELPERS ***********************/

func makeFreshQuery(t string, o string, d string) url.Values {
	q := url.Values{}
	q.Set("t", t)
	if o != "" {
		q.Set("o", o)
	}
	if d != "" {
		q.Set("d", d)
	}
	return q
}

func makeRowNum(q url.Values, rownum int64) Entry {
	q.Set("n", Int64toa(rownum))
	link := q.Encode()
	q.Del("n")
	return escape(Int64toa(rownum), link)
}

func makeValuesPointers(columns []string) ([]interface{}, []interface{}) {
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for i, _ := range columns {
		valuePtrs[i] = &values[i]
	}
	return values, valuePtrs
}

func makeRecords(rows *sql.Rows, t string, primary string, offset int64, q url.Values) ([][]Entry, int64) {

	//q, err := url.ParseQuery(original.Encode());checkY(err)   // brute force to preserve original
	columns, err := rows.Columns()
	checkY(err)
	values, valuePtrs := makeValuesPointers(columns)
	records := [][]Entry{}
	rownum := offset
	for rows.Next() {
		rownum = rownum + 1
		row := []Entry{}
		row = append(row, makeRowNum(q, rownum))
		err = rows.Scan(valuePtrs...)
		checkY(err)

		for i, c := range columns {
			nv := getNullString(values[i])
			row = append(row, makeEntry(nv, t, c, primary, q))
		}
		records = append(records, row)
	}
	return records, rownum
}
