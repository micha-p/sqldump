package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

func dumpIt(w http.ResponseWriter, cred Access, db string, t string, o string, d string, n string, k string, v string) {

	q := url.Values{}
	trail := []Entry{}
	trail = append(trail, Entry{"/", cred.Host})
	var query string

	if db == "" {
		dumpHome(w, cred, trail, "/logout")
		return
	} else {
		q.Add("db", db)
		trail = append(trail, Entry{Link: "?" + q.Encode(), Label: db})
	}

	if t == "" {
		dumpTables(w, db, cred, trail, "?"+q.Encode())
		return
	} else {
		q.Add("t", t)
		trail = append(trail, Entry{Link: "?" + q.Encode(), Label: t})
	}

	nnumber, err := regexp.MatchString("^ *\\d+ *$", n)
	checkY(err)

	if k != "" && v != "" && k == getPrimary(cred, db, t) {
		q.Add("k", k)
		q.Add("v", v)
		trail = append(trail, Entry{Link: "/?" + q.Encode(), Label: k + " (ID)"})
		query = "select * from " + template.HTMLEscapeString(t) + " where " + template.HTMLEscapeString(k) + "=" + template.HTMLEscapeString(v)
		dumpKeyValue(w, db, t, k, v, cred, trail, "?"+q.Encode(), query)
	} else if nnumber {
		if o != "" {
			q.Add("o", o)
			query = "select * from " + template.HTMLEscapeString(t) + " order by " + template.HTMLEscapeString(o)
	        if d == "" {
				trail = append(trail, Entry{Link: "/?" + q.Encode(), Label: o + "&uarr;"})
			} else {
				q.Add("d", d)
				trail = append(trail, Entry{Link: "/?" + q.Encode(), Label: o + "&darr;"})
				query = query + " desc"
			} 
		} else {
			query = "select * from " + template.HTMLEscapeString(t)
		}
		dumpFields(w, db, t, o, d, n, cred, trail, "?"+q.Encode(), query)
	} else {
		if o != "" {
			q.Add("o", o)
			query = "select * from " + template.HTMLEscapeString(t) + " order by " + template.HTMLEscapeString(o)
			if d == "" {
				trail = append(trail, Entry{Link: "/?" + q.Encode(), Label: o + "&uarr;"})
			} else {
				q.Add("d", d)
				trail = append(trail, Entry{Link: "/?" + q.Encode(), Label: o + "&darr;"})
				query = query + " desc"
			}
		} else {
			query = "select * from " + template.HTMLEscapeString(t)
		}
		re := regexp.MustCompile("^ *(\\d+) *- *(\\d+) *$")
		limits := re.FindStringSubmatch(n)

		if len(limits) == 3 {
			startint, err := strconv.Atoi(limits[1])
			checkY(err)
			startint = maxI(startint, 1)
			endint, err := strconv.Atoi(limits[2])
			checkY(err)
			maxint, err := strconv.Atoi(getCount(cred, db, t))
			checkY(err)
			endint = minI(endint, maxint)
			query = query + " limit " + strconv.Itoa(1+endint-startint) + " offset " + strconv.Itoa(startint-1)
			dumpRange(w, db, t, o, d, startint, endint, maxint, cred, trail, "?"+q.Encode(), query)
		} else {
			dumpRows(w, db, t, o, d, cred, trail, "?"+q.Encode(), query)
		}
	}
}

// Shows selection of databases at top level
func dumpHome(w http.ResponseWriter, cred Access, trail []Entry, back string) {

	q := url.Values{}
	rows := getRows(cred, "", "show databases")
	defer rows.Close()

	menu := []Entry{}
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
	tableOut(w, cred, "", "", back, head, records, trail, menu)
}

//  Dump all tables of a database
func dumpTables(w http.ResponseWriter, db string, cred Access, trail []Entry, back string) {

	q := url.Values{}
	q.Add("db", db)
	rows := getRows(cred, db, "show tables")
	defer rows.Close()

	menu := []Entry{}
	records := [][]string{}
	head := []string{"#", "Table", "Rows"}

	var n int = 1
	for rows.Next() {
		var field string
		var nrows string
		rows.Scan(&field)
		if db == "information_schema" {
			nrows = "?"
		} else {
			nrows = getCount(cred, db, field)
		}
		q.Set("t", field)
		link := q.Encode()
		row := []string{href(link, strconv.Itoa(n)), href(link, field), nrows}
		records = append(records, row)
		n = n + 1
	}
	tableOut(w, cred, db, "", back, head, records, trail, menu)
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

func createHead(db string, t string, o string, d string, primary string, columns []string, q url.Values) []string {
	head := []string{}

	if showNumsBool(primary, o) {
		head = append(head, "#")
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
			if d=="" {
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

func dumpRows(w http.ResponseWriter, db string, t string, o string, d string, cred Access, trail []Entry, back string, query string) {

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("action", "add")
	linkinsert := "/?" + q.Encode()
	q.Set("action", "subset")
	linkselect := "/?" + q.Encode()
	q.Set("action", "info")
	linkinfo := "?" + q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, Entry{linkselect, "?"})
	menu = append(menu, Entry{linkinfo, "i"})
	menu = append(menu, Entry{linkinsert, "+"})

	primary := getPrimary(cred, db, t)
	rows := getRows(cred, db, query)
	defer rows.Close()
	columns, err := rows.Columns()
	checkY(err)
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for i, _ := range columns {
		valuePtrs[i] = &values[i]
	}

	head := createHead(db, t, o, d, primary, columns, q)
	
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

	limitstring := "1-" + strconv.Itoa(rownum-1)
	q.Set("n", limitstring)
	link := "?" + q.Encode()
	tableOutRows(w, cred, db, t, o, d, limitstring, link, link, back, head, records, trail, menu)
}

func dumpRange(w http.ResponseWriter, db string, t string, o string, d string, start int, end int, max int, cred Access, trail []Entry, back string, query string) {

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("action", "add")
	linkinsert := "/?" + q.Encode()
	q.Set("action", "subset")
	linkselect := "/?" + q.Encode()
	q.Set("action", "info")
	linkinfo := "?" + q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, Entry{linkselect, "?"})
	menu = append(menu, Entry{linkinfo, "i"})
	menu = append(menu, Entry{linkinsert, "+"})

	limitstring := strconv.Itoa(start) + "-" + strconv.Itoa(end)
	q.Add("n", limitstring)

	primary := getPrimary(cred, db, t)
	rows := getRows(cred, db, query)
	defer rows.Close()
	columns, err := rows.Columns()
	checkY(err)
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for i, _ := range columns {
		valuePtrs[i] = &values[i]
	}

	head := createHead(db, t, o, d, "", columns, q)

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
	tableOutRows(w, cred, db, t, o, d, limitstring, linkleft, linkright, back, head, records, trail, menu)
}

// Dump all fields of a record, one column per line
func dumpFields(w http.ResponseWriter, db string, t string, o string, d string, n string, cred Access, trail []Entry, back string, query string) {

	rows := getRows(cred, db, query)
	defer rows.Close()
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

	v.Add("action", "add")
	linkinsert := "/?" + v.Encode()
	v.Set("action", "info")
	linkinfo := "?" + v.Encode()
	v.Del("action")

	menu := []Entry{}
	menu = append(menu, Entry{linkinfo, "i"})
	menu = append(menu, Entry{linkinsert, "+"})

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

	tableOutFields(w, cred, db, t, o, d, n, linkleft, linkright, back, head, records, trail, menu)
}

func dumpKeyValue(w http.ResponseWriter, db string, t string, k string, v string, cred Access, trail []Entry, back string, query string) {

	rows := getRows(cred, db, query)
	defer rows.Close()
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

	var iter int = 1
rowLoop:
	for rows.Next() {

		// just one row
		if iter == 1 {
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

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)

	q.Add("action", "add")
	linkinsert := "/?" + q.Encode()
	q.Set("action", "info")
	linkinfo := "?" + q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, Entry{linkinfo, "i"})
	menu = append(menu, Entry{linkinsert, "+"})

	/*
	 mysql> select number from unique_field where number>12 order by number limit 1;
	 mysql> select number from unique_field where number<123 order by number desc limit 1;
	 mysql> select content from unique_field where content > "hu" order by content limit 1;
	*/

	next := getSingle(cred, db, "select "+k+" from "+t+" where "+k+">"+v+" order by "+k+" limit 1")
	if next == "<nil>" {
		next = v
	}
	prev := getSingle(cred, db, "select "+k+" from "+t+" where "+k+"<"+v+" order by "+k+" desc limit 1")
	if prev == "<nil>" {
		prev = v
	}
	q.Set("k", k)
	q.Set("v", prev)
	linkleft := "?" + q.Encode()
	q.Set("v", next)
	linkright := "?" + q.Encode()

	tableOutFields(w, cred, db, t, "", "", v, linkleft, linkright, back, head, records, trail, menu)
}
