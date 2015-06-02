package main

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

func dumpIt(w http.ResponseWriter, cred Access, db string, t string, o string, d string, n string, k string, v string, where string) {

	var query string
	nnumber, err := regexp.MatchString("^ *\\d+ *$", n)
	checkY(err)
	re := regexp.MustCompile("^ *(\\d+) *- *(\\d+) *$")
	limits := re.FindStringSubmatch(n)

	if db == "" {
		dumpHome(w, cred)
		return
	} else if t == "" {
		dumpTables(w, db, cred)
	} else if k != "" && v != "" && k == getPrimary(cred, db, t) {
		query = "select * from " + t + " where " + k + "=" + v
		dumpKeyValue(w, db, t, k, v, cred, query)
	} else if nnumber {
		if o != "" {
			query = "select * from " + t + " order by " + o
			if d != "" {
				query = query + " desc"
			}
		} else {
			query = "select * from " + t
		}
		dumpFields(w, db, t, o, d, n, cred, query)
	} else if len(limits) == 3 {
		startint, err := strconv.Atoi(limits[1])
		checkY(err)
		startint = maxI(startint, 1)
		endint, err := strconv.Atoi(limits[2])
		checkY(err)
		maxint, err := strconv.Atoi(getCount(cred, db, t))
		checkY(err)
		endint = minI(endint, maxint)
		query = "select * from " + t + " limit " + strconv.Itoa(1+endint-startint) + " offset " + strconv.Itoa(startint-1)
		if o != "" {
			query = "select t.* from (" + query + ") t order by " + o
			if d != "" {
				query = query + " desc"
			}
		}
		dumpRange(w, db, t, o, d, startint, endint, maxint, cred, query)
	} else {
		query = "select * from " + t
		if o != "" {
			query = query + " order by " + o
			if d != "" {
				query = query + " desc"
			}
		}
		dumpRows(w, db, t, o, d, cred, query, "")
	}
}

// Shows selection of databases at top level
func dumpHome(w http.ResponseWriter, cred Access) {

	q := url.Values{}
	rows, err := getRows(cred, "", "show databases")
	checkY(err)
	defer rows.Close()

	records := [][]string{}
	head := []string{"#", "Database"}
	var n int = 1
	for rows.Next() {
		var field string
		rows.Scan(&field)
		if EXPERTFLAG || INFOFLAG || field != "information_schema" {
			q.Set("db", field)
			link := q.Encode()
			row := []string{href(link, strconv.Itoa(n)), href(link, field)}
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

	records := [][]string{}
	head := []string{"#", "Table", "Rows"}

	var n int = 1
	for rows.Next() {
		var field string
		var nrows string
		rows.Scan(&field)
		nrows = getCount(cred, db, field)

		q.Set("t", field)
		link := q.Encode()
		row := []string{href(link, strconv.Itoa(n)), href(link, field), nrows}
		records = append(records, row)
		n = n + 1
	}
	tableOutSimple(w, cred, db, "", head, records, []Entry{})
}

// http://stackoverflow.com/questions/17845619/how-to-call-the-scan-variadic-function-in-golang-using-reflection/17885636#17885636
// http://blog.golang.org/laws-of-reflection

func dumpValue(val interface{}) string {

	var r string
	b, ok := val.([]byte)

	if ok {
		r = string(b)
		if r == "" {
			r = "NULL"
		}
	} else {
		r = fmt.Sprint(val)
		if r == "" {
			r = "NULL NOT OK"
		}
	}
	return r
}

func showNumsBool(primary string, o string) bool {
	if primary == "" || o == "" || (o != "" && o != primary) {
		return true
	} else {
		return false
	}
}

func createHead(db string, t string, o string, d string, n string, primary string, columns []string, q url.Values) []string {
	root := url.Values{}
	head := []string{}

	if showNumsBool(primary, o) {
		root.Add("db", db)
		root.Add("t", t)
		root.Add("n", n)
		head = append(head, href(root.Encode(), "#"))
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
				head = append(head, href(q.Encode(), titlestring+"&uarr;"))
			} else {
				q.Del("d")
				head = append(head, href(q.Encode(), titlestring+"&darr;"))
			}
		} else {
			q.Set("o", title)
			q.Del("d")
			head = append(head, href(q.Encode(), titlestring))
		}
	}
	return head
}

func dumpRows(w http.ResponseWriter, db string, t string, o string, d string, cred Access, query string, where string) {

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("action", "ADD")
	linkinsert := "/?" + q.Encode()
	q.Set("action", "SUBSET")
	linkselect := "/?" + q.Encode()
	q.Set("action", "INFO")
	linkinfo := "?" + q.Encode()
	if where != "" {
		q.Add("where", where)
	}
	q.Set("action", "DELETE")
	linkdelete := "?" + q.Encode()
	q.Set("action", "UPDATE")
	linkupdate := "?" + q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, Entry{linkselect, "?"})
	menu = append(menu, Entry{linkinsert, "+"})
	if where != "" {
		menu = append(menu, Entry{linkupdate, "~"})
		menu = append(menu, Entry{linkdelete, "-"})
	}
	menu = append(menu, Entry{linkinfo, "i"})

	rows, err := getRows(cred, db, query)
	if err != nil {
		shipError(w, cred, db, t, query, err)
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

	head := createHead(db, t, o, d, "", primary, columns, q)

	if o != "" {
		q.Set("o", o)
	} else {
		q.Del("o")
	}
	if d != "" {
		q.Set("d", d)
	} else {
		q.Del("d")
	}

	records := [][]string{}
	rownum := 1
	for rows.Next() {

		row := []string{}
		if showNumsBool(primary, o) {
			q.Set("n", strconv.Itoa(rownum))
			row = append(row, href(q.Encode(), strconv.Itoa(rownum)))
		}

		err = rows.Scan(valuePtrs...)
		checkY(err)

		for i, _ := range columns {
			v := dumpValue(values[i])
			if columns[i] == primary {
				q.Del("o")
				q.Del("d")
				q.Del("n")
				q.Set("k", primary)
				q.Set("v", v)
				row = append(row, href(q.Encode(), v))
			} else {
				row = append(row, v)
			}
		}

		records = append(records, row)
		rownum = rownum + 1
	}

	var limitstring, link string
	if where == "" {
		limitstring = "1-" + strconv.Itoa(rownum-1)
		q.Set("n", limitstring)
		link = "?" + q.Encode()
	} else {
		limitstring = ""
		link = ""
	}
	tableOutRows(w, cred, db, t, o, d, limitstring, link, link, head, records, menu, where)
}

func dumpRange(w http.ResponseWriter, db string, t string, o string, d string, start int, end int, max int, cred Access, query string) {

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("action", "ADD")
	linkinsert := "/?" + q.Encode()
	q.Set("action", "SUBSET")
	linkselect := "/?" + q.Encode()
	q.Set("action", "INFO")
	linkinfo := "?" + q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, Entry{linkselect, "?"})
	menu = append(menu, Entry{linkinsert, "+"})
	menu = append(menu, Entry{linkinfo, "i"})

	limitstring := strconv.Itoa(start) + "-" + strconv.Itoa(end)
	q.Add("n", limitstring)

	rows, err := getRows(cred, db, query)
	if err != nil {
		shipError(w, cred, db, t, query, err)
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

	records := [][]string{}
	rowrange := 1 + end - start
	rownum := start
	for rows.Next() && rownum <= end {

		if o != "" {
			q.Set("o", o)
		}
		if d != "" {
			q.Set("d", d)
		}
		row := []string{}
		q.Set("n", strconv.Itoa(rownum))
		row = append(row, href(q.Encode(), strconv.Itoa(rownum)))

		err = rows.Scan(valuePtrs...)
		checkY(err)

		for i, _ := range columns {
			v := dumpValue(values[i])
			if columns[i] == primary {
				q.Del("o")
				q.Del("d")
				q.Del("n")
				q.Set("k", primary)
				q.Set("v", v)
				row = append(row, href(q.Encode(), v))
			} else {
				row = append(row, v)
			}
		}

		records = append(records, row)
		rownum = rownum + 1
	}

	left := maxI(start-rowrange, 1)
	right := minI(end+rowrange, max)
	q.Set("n", strconv.Itoa(left)+"-"+strconv.Itoa(left+rowrange-1))
	linkleft := "?" + q.Encode()
	q.Set("n", strconv.Itoa(1+right-rowrange)+"-"+strconv.Itoa(right))
	linkright := "?" + q.Encode()
	tableOutRows(w, cred, db, t, o, d, limitstring, linkleft, linkright, head, records, menu, "")
}

// Dump all fields of a record, one column per line
func dumpFields(w http.ResponseWriter, db string, t string, o string, d string, n string, cred Access, query string) {

	rows, err := getRows(cred, db, query)
	if err != nil {
		shipError(w, cred, db, t, query, err)
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

	head := []string{"#", "Column", "Data"}
	records := [][]string{}

	rec, err := strconv.Atoi(n)
	checkY(err)
	var iter int = 1
rowLoop:
	for rows.Next() {

		// unfortunately we have to iterate up to row of interest
		if iter == rec {
			err = rows.Scan(valuePtrs...)
			checkY(err)
			for i, _ := range columns {
				var row []string
				var colstring string
				if columns[i] == primary {
					colstring = primary + " (ID)"
				} else {
					colstring = columns[i]
				}
				row = []string{strconv.Itoa(i + 1), colstring, dumpValue(values[i])}
				records = append(records, row)
			}
			break rowLoop
		}
		iter = iter + 1
	}

	v := url.Values{}
	v.Add("db", db)
	v.Add("t", t)

	v.Add("action", "ADD")
	linkinsert := "/?" + v.Encode()
	v.Set("action", "INFO")
	linkinfo := "?" + v.Encode()
	v.Del("action")

	menu := []Entry{}
	menu = append(menu, Entry{linkinsert, "+"})
	menu = append(menu, Entry{linkinfo, "i"})

	nint, err := strconv.Atoi(n)
	checkY(err)
	nmax, err := strconv.Atoi(getCount(cred, db, t))
	checkY(err)
	left := strconv.Itoa(maxI(nint-1, 1))
	right := strconv.Itoa(minI(nint+1, nmax))

	if o != "" {
		v.Set("o", o)
	}
	if d != "" {
		v.Set("d", d)
	}
	v.Set("n", left)
	linkleft := "?" + v.Encode()
	v.Set("n", right)
	linkright := "?" + v.Encode()

	tableOutFields(w, cred, db, t, o, d, "", n, linkleft, linkright, head, records, menu)
}

func dumpKeyValue(w http.ResponseWriter, db string, t string, k string, v string, cred Access, query string) {

	rows, err := getRows(cred, db, query)
	if err != nil {
		shipError(w, cred, db, t, query, err)
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

	head := []string{"#", "Column", "Data"}
	records := [][]string{}

	rows.Next() // just one row
	err = rows.Scan(valuePtrs...)
	checkY(err)

	for i, _ := range columns {
		var row []string
		var colstring string
		if columns[i] == primary {
			colstring = primary + " (ID)"
		} else {
			colstring = columns[i]
		}
		row = []string{strconv.Itoa(i + 1), colstring, dumpValue(values[i])}
		records = append(records, row)
	}

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("action", "ADD")
	linkinsert := "/?" + q.Encode()
	q.Set("action", "INFO")
	linkinfo := "?" + q.Encode()
	q.Add("k", k)
	q.Add("v", v)
	q.Set("action", "REMOVE")
	linkremove := "?" + q.Encode()
	q.Set("action", "EDIT")
	linkedit := "?" + q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, Entry{linkinsert, "+"})
	menu = append(menu, Entry{linkedit, "~"})
	menu = append(menu, Entry{linkremove, "-"})
	menu = append(menu, Entry{linkinfo, "i"})

	next := getSingleValue(cred, db, "select "+k+" from "+t+" where "+k+">"+v+" order by "+k+" limit 1")
	if next == "NULL" {
		next = v
	}
	prev := getSingleValue(cred, db, "select "+k+" from "+t+" where "+k+"<"+v+" order by "+k+" desc limit 1")
	if prev == "NULL" {
		prev = v
	}

	q.Set("v", prev)
	linkleft := "?" + q.Encode()
	q.Set("v", next)
	linkright := "?" + q.Encode()

	tableOutFields(w, cred, db, t, "", "", k, v, linkleft, linkright, head, records, menu)
}
