package main

import (
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

func dumpSelection(w http.ResponseWriter, cred Access, db string, t string, o string, d string, n string, k string, v string) {

	query := "select * from `" + t + "`"
	nnumber := regexp.MustCompile("^ *(\\d+) *$").FindString(n)
	limits := regexp.MustCompile("^ *(\\d+) *- *(\\d+) *$").FindStringSubmatch(n)

	if k != "" && v != "" && k == getPrimary(cred, db, t) {
		query = query + " where `" + k + "` =" + v
		dumpKeyValue(w, db, t, k, v, cred, query)
	} else if nnumber != "" {
		if o != "" {
			query = query + " order by `" + o + "`"
			if d != "" {
				query = query + " desc"
			}
		}
		nint, err := strconv.Atoi(nnumber)
		checkY(err)
		maxint, err := strconv.Atoi(getCount(cred, db, t))
		checkY(err)
		nint = minI(nint, maxint)
		query = query + " limit 1 offset " + strconv.Itoa(nint-1)
		dumpFields(w, db, t, o, d, nnumber, nint, maxint, cred, query)
	} else if len(limits) == 3 {
		startint, err := strconv.Atoi(limits[1])
		checkY(err)
		startint = maxI(startint, 1)
		endint, err := strconv.Atoi(limits[2])
		checkY(err)
		maxint, err := strconv.Atoi(getCount(cred, db, t))
		checkY(err)
		endint = minI(endint, maxint)
		query = query + " limit " + strconv.Itoa(1+endint-startint) + " offset " + strconv.Itoa(startint-1)
		if o != "" {
			query = "select t.* from (" + query + ") t order by `" + o + "`"
			if d != "" {
				query = query + " desc"
			}
		}
		dumpRange(w, db, t, o, d, startint, endint, maxint, cred, query)
	} else {
		dumpRows(w, db, t, o, d, cred, query, url.Values{})
	}
}


func dumpRows(w http.ResponseWriter, db string, t string, o string, d string, cred Access, query string, where url.Values) {

	wherestring := ""
	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	if o != "" {
		q.Add("o", o)
		query = query + " order by `" + o + "`"
		if d != "" {
			q.Add("d", d)
			query = query + " desc"
		}
	}

	rows, err := getRows(cred, db, query)
	if err != nil {
		checkErrorPage(w, cred, db, t, query, err)
		return
	} else {
		defer rows.Close()
	}

	primary := getPrimary(cred, db, t)
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
	rownum := 1
	for rows.Next() {

		row := []Entry{}
		q.Del("k")
		q.Del("v")
		q.Set("d",d)
		q.Set("o",o)
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
					row = append(row, escape(v, "")) // TODO add single where clause
				}
			} else {
				row = append(row, escapeNull())
			}
		}

		records = append(records, row)
		rownum = rownum + 1
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

	menu := []Entry{}
	var limitstring string
	var linkleft,linkright Entry
	if len(where) > 0 {
		limitstring = ""
		wherestring = WhereSelect2Pretty(where, getColumnInfo(cred, db, t))
		where.Add("db", db)
		where.Add("t", t)
		where.Set("action", "QUERY")
		linkselect = q.Encode()
		where.Set("action", "ADD")
		linkinsert = q.Encode()
		where.Set("action", "UPDATEFORM")
		linkupdate = where.Encode()
		where.Set("action", "DELETE")
		linkdelete := where.Encode()
		head = createHead(db, t, o, d, "", primary, columns, where)
		menu = append(menu, escape("?",linkselect))
		menu = append(menu, escape("+",linkinsert))
		menu = append(menu, escape("~",linkupdate))
		menu = append(menu, escape("-",linkdelete))
		menu = append(menu, escape("i",linkinfo))
	} else {
		limitstring = "1-" + strconv.Itoa(rownum-1)
		q.Del("n")
		linkleft = escape("<",q.Encode())
		linkright = escape(">",q.Encode())
		menu = append(menu, escape("?",linkselect))
		menu = append(menu, escape("+",linkinsert))
		menu = append(menu, escape("~",linkupdate))
		menu = append(menu, escape("-",linkdeleteF))
		menu = append(menu, escape("i",linkinfo))
	}
	tableOutRows(w, cred, db, t, primary, o, d, limitstring, linkleft, linkright, head, records, menu, wherestring, where)
}

func dumpRange(w http.ResponseWriter, db string, t string, o string, d string, start int, end int, max int, cred Access, query string) {

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
	menu = append(menu, escape("?",linkselect))
	menu = append(menu, escape("+",linkinsert))
	menu = append(menu, escape("i",linkinfo))

	limitstring := strconv.Itoa(start) + "-" + strconv.Itoa(end)
	q.Add("n", limitstring)

	rows, err := getRows(cred, db, query)
	if err != nil {
		checkErrorPage(w, cred, db, t, query, err)
		return
	} else {
		defer rows.Close()
	}

	primary := getPrimary(cred, db, t)
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
	rowrange := 1 + end - start
	rownum := start
	for rows.Next() && rownum <= end {

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
					row = append(row, escape(v, ""))
				}
			} else {
				row = append(row, escapeNull())
			}
		}

		records = append(records, row)
		rownum = rownum + 1
	}

	left := maxI(start-rowrange, 1)
	right := minI(end+rowrange, max)
	q.Set("n", strconv.Itoa(left)+"-"+strconv.Itoa(left+rowrange-1))
	linkleft := escape("<",q.Encode())
	q.Set("n", strconv.Itoa(1+right-rowrange)+"-"+strconv.Itoa(right))
	linkright := escape(">",q.Encode())
	tableOutRows(w, cred, db, t, primary, o, d, limitstring, linkleft, linkright, head, records, menu, "", url.Values{})
}

// Dump all fields of a record, one column per line
func dumpFields(w http.ResponseWriter, db string, t string, o string, d string, n string, nint int, nmax int, cred Access, query string) {

	rows, err := getRows(cred, db, query)
	checkY(err)
	vmap := getNullStringMap(rows)
	head := []Entry{escape("#"), escape("Column"), escape("Data")}
	records := [][]Entry{}

	i := 1
	for f, nv := range vmap { // TODO should be range cols
		v := nv.String
		var row []Entry
		row = []Entry{escape(strconv.Itoa(i), ""), escape(f, ""), escape(v, "")}
		records = append(records, row)
		i = i + 1
	}

	v := url.Values{}
	v.Add("db", db)
	v.Add("t", t)
	v.Add("action", "ADD")
	linkinsert := v.Encode()
	v.Set("action", "INFO")
	linkinfo := v.Encode()
	v.Del("action")

	menu := []Entry{}
	menu = append(menu, escape("+",linkinsert))
	menu = append(menu, escape("i",linkinfo))

	left := strconv.Itoa(maxI(nint-1, 1))
	right := strconv.Itoa(minI(nint+1, nmax))

	if o != "" {
		v.Set("o", o)
	}
	if d != "" {
		v.Set("d", d)
	}
	v.Set("n", left)
	linkleft := escape("<",v.Encode())
	v.Set("n", right)
	linkright := escape(">",v.Encode())

	tableOutFields(w, cred, db, t, "", o, d, "", n, linkleft, linkright, head, records, menu)
}

func dumpKeyValue(w http.ResponseWriter, db string, t string, k string, v string, cred Access, query string) {

	rows, err := getRows(cred, db, query)
	checkY(err)
	vmap := getNullStringMap(rows)
	primary := getPrimary(cred, db, t)
	head := []Entry{escape("#"), escape("Column"), escape("Data")}
	records := [][]Entry{}

	i := 1
	for f, nv := range vmap { // TODO should be range cols
		v := nv.String
		var row []Entry
		row = []Entry{escape(strconv.Itoa(i), ""), escape(f, ""), escape(v, "")}
		records = append(records, row)
		i = i + 1
	}

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("action", "ADD")
	linkinsert := q.Encode()
	q.Set("action", "INFO")
	linkinfo := q.Encode()
	q.Add("k", k)
	q.Add("v", v)
	q.Set("action", "DELETEPRI")
	linkDELETEPRI := q.Encode()
	q.Set("action", "EDITFORM")
	linkedit := q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, escape("+",linkinsert))
	menu = append(menu, escape("~",linkedit))
	menu = append(menu, escape("-",linkDELETEPRI))
	menu = append(menu, escape("i",linkinfo))

	next, err := getSingleValue(cred, db, "select `"+k+"` from `"+t+"` where `"+k+"` > "+v+" order by `"+k+"` limit 1")
	if err == nil {
		q.Set("v", next)
	} else {
		q.Set("v", v)
	}
	linkright := escape(">",q.Encode())
	prev, err := getSingleValue(cred, db, "select `"+k+"` from `"+t+"` where `"+k+"` < "+v+" order by `"+k+"` desc limit 1")
	if err == nil {
		q.Set("v", prev)
	} else {
		q.Set("v", v)
	}
	linkleft := escape("<",q.Encode())
	tableOutFields(w, cred, db, t, primary, k, "", k, v, linkleft, linkright, head, records, menu)
}
