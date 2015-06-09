package main

import (
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

func dumpIt(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, o string, d string, n string, k string, v string) {

	if db == "" {
		dumpHome(w, cred)
		return
	} else if t == "" {
		dumpTables(w, db, cred)
	} else {
		dumpSelection(w, cred, db, t, o, d, n, k, v)
	}
}

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

// Shows selection of databases at top level
func dumpHome(w http.ResponseWriter, cred Access) {

	q := url.Values{}
	rows, err := getRows(cred, "", "show databases")
	checkY(err)
	defer rows.Close()

	records := [][]Entry{}
	head := []Entry{{"#", ""}, {"Database", ""}}
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
	tableOutSimple(w, cred, "", "", head, records, []Entry{})
}

//  Dump all tables of a database
func dumpTables(w http.ResponseWriter, db string, cred Access) {

	q := url.Values{}
	q.Add("db", db)
	rows, err := getRows(cred, db, "show tables")
	checkY(err)
	defer rows.Close()

	records := [][]Entry{}
	head := []Entry{{"#", ""}, {"Table", ""}, {"Rows", ""}}

	var n int = 1
	for rows.Next() {
		var field string
		var nrows string
		rows.Scan(&field)
		nrows = getCount(cred, db, field)

		q.Set("t", field)
		link := q.Encode()
		row := []Entry{escape(strconv.Itoa(n), link), escape(field, link), escape(nrows, "")}
		records = append(records, row)
		n = n + 1
	}
	tableOutSimple(w, cred, db, "", head, records, []Entry{})
}

func showNumsBool(primary string, o string) bool {
	if primary == "" || o == "" || (o != "" && o != primary) {
		return true
	} else {
		return false
	}
}

func createHead(db string, t string, o string, d string, n string, primary string, columns []string, q url.Values) []Entry {
	root := url.Values{}
	head := []Entry{}

	if showNumsBool(primary, o) {
		root.Add("db", db)
		root.Add("t", t)
		root.Add("n", n)
		head = append(head, Entry{"#", root.Encode()})
	}
	for _, title := range columns {
		var titlestring string
		if primary == title {
			titlestring = title + " (ID)"
			if o == title {
				titlestring = "# " + titlestring
			}
		} else {
			titlestring = title
		}

		if o == title {
			q.Set("o", title)
			if d == "" {
				q.Set("d", "1")
				head = append(head, Entry{Link: q.Encode(), Text: titlestring + "&uarr;"})
			} else {
				q.Del("d")
				head = append(head, Entry{Link: q.Encode(), Text: titlestring + "&darr;"})
			}
		} else {
			q.Set("o", title)
			q.Del("d")
			head = append(head, Entry{Link: q.Encode(), Text: titlestring})
		}
	}
	return head
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

	q.Add("action", "ADD")
	linkinsert := q.Encode()
	q.Set("action", "QUERY")
	linkselect := q.Encode()
	q.Set("action", "INFO")
	linkinfo := q.Encode()
	q.Set("action", "DELETEFORM")
	linkdeleteF := q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, Entry{Link: linkselect, Text: "?"})
	menu = append(menu, Entry{Link: linkinsert, Text: "+"})

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
		if showNumsBool(primary, o) {
			q.Del("k")
			q.Del("v")
			q.Set("n", strconv.Itoa(rownum))
			row = append(row, escape(strconv.Itoa(rownum), q.Encode()))
		}

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

	var limitstring, link string
	if len(where) > 0 {
		limitstring = ""
		link = ""
		wherestring = WhereSelect2Pretty(where, getColumnInfo(cred, db, t))
		where.Add("db", db)
		where.Add("t", t)
		where.Set("action", "DELETE")
		linkdelete := where.Encode()
		where.Set("action", "UPDATE")
		linkupdate := where.Encode()
		where.Set("action", "SELECT")
		head = createHead(db, t, o, d, "", primary, columns, where)
		menu = append(menu, Entry{Link: linkupdate, Text: "~"})
		menu = append(menu, Entry{Link: linkdelete, Text: "-"})
	} else {
		limitstring = "1-" + strconv.Itoa(rownum-1)
		q.Set("n", limitstring)
		link = q.Encode()
		menu = append(menu, Entry{Link: linkdeleteF, Text: "-"})
	}
	menu = append(menu, Entry{Link: linkinfo, Text: "i"})

	tableOutRows(w, cred, db, t, o, d, limitstring, link, link, head, records, menu, wherestring, where)
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
	menu = append(menu, Entry{Link: linkselect, Text: "?"})
	menu = append(menu, Entry{Link: linkinsert, Text: "+"})
	menu = append(menu, Entry{Link: linkinfo, Text: "i"})

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
	linkleft := q.Encode()
	q.Set("n", strconv.Itoa(1+right-rowrange)+"-"+strconv.Itoa(right))
	linkright := q.Encode()
	tableOutRows(w, cred, db, t, o, d, limitstring, linkleft, linkright, head, records, menu, "", url.Values{})
}

// Dump all fields of a record, one column per line
func dumpFields(w http.ResponseWriter, db string, t string, o string, d string, n string, nint int, nmax int, cred Access, query string) {

	rows, err := getRows(cred, db, query)
	checkY(err)
	vmap := getNullStringMap(w, db, t, cred, rows)
	head := []Entry{{"#", ""}, {"Column", ""}, {"Data", ""}}
	records := [][]Entry{}

	i := 1
	for f, nv := range vmap {
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
	menu = append(menu, Entry{"+", linkinsert})
	menu = append(menu, Entry{"i", linkinfo})

	left := strconv.Itoa(maxI(nint-1, 1))
	right := strconv.Itoa(minI(nint+1, nmax))

	if o != "" {
		v.Set("o", o)
	}
	if d != "" {
		v.Set("d", d)
	}
	v.Set("n", left)
	linkleft := v.Encode()
	v.Set("n", right)
	linkright := v.Encode()

	tableOutFields(w, cred, db, t, o, d, "", n, linkleft, linkright, head, records, menu)
}

func dumpKeyValue(w http.ResponseWriter, db string, t string, k string, v string, cred Access, query string) {

	rows, err := getRows(cred, db, query)
	checkY(err)
	vmap := getNullStringMap(w, db, t, cred, rows)
	primary := getPrimary(cred, db, t)
	head := []Entry{{"#", ""}, {"Column", ""}, {"Data", ""}}
	records := [][]Entry{}

	i := 1
	for f, nv := range vmap {
		v := nv.String
		var row []Entry
		if f == primary {
			f = f + " (ID)"
		}
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
	q.Set("action", "REMOVE")
	linkremove := q.Encode()
	q.Set("action", "EDIT")
	linkedit := q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, Entry{Link: linkinsert, Text: "+"})
	menu = append(menu, Entry{Link: linkedit, Text: "~"})
	menu = append(menu, Entry{Link: linkremove, Text: "-"})
	menu = append(menu, Entry{Link: linkinfo, Text: "i"})

	next := getSingleValue(cred, db, "select `"+k+"` from `"+t+"` where `"+k+"` > "+v+" order by `"+k+"` limit 1")
	if next == "NULL" {
		next = v
	}
	prev := getSingleValue(cred, db, "select `"+k+"` from `"+t+"` where `"+k+"` < "+v+" order by `"+k+"` desc limit 1")
	if prev == "NULL" {
		prev = v
	}

	q.Set("v", prev)
	linkleft := q.Encode()
	q.Set("v", next)
	linkright := q.Encode()

	tableOutFields(w, cred, db, t, "", "", k, v, linkleft, linkright, head, records, menu)
}
