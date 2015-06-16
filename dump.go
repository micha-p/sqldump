package main

import (
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

func dumpSelection(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, o string, d string, n string, k string, v string) {

	query := sqlStar(t)
	
	// we have to retrieve columns for getting where clauses => TODO: w1=column,c1,v1=value,s1=column
	cols := getCols(cred, db, t)
	wclauses, _, whereQ := collectClauses(r, cols)
	
	if len(wclauses)>0 {
		query = "select t.* from (" + query + sqlWhereClauses(wclauses) + ") t "
	}
	if o != "" {
		query = query + sqlOrder(o,d)
	}
	if n != "" {
		maxint, err := strconv.Atoi(getCount(cred, db, t))
		checkY(err)
		singlenumber := regexp.MustCompile("^ *(\\d+) *$").FindString(n)
		limits := regexp.MustCompile("^ *(\\d+) *- *(\\d+) *$").FindStringSubmatch(n)
		
		if singlenumber != "" {
			nint, err := strconv.Atoi(singlenumber)
			checkY(err)
			nint = minI(nint, maxint)
			query = query + sqlLimit(1,nint)
			// maxint depends on query!
			// TODO: 
			// for maxint check if rows.next available!
			dumpFields(w, cred, db, t, o, d, singlenumber, nint, maxint, query,whereQ)
		} else if len(limits) == 3 {
			startint, err := strconv.Atoi(limits[1])
			checkY(err)
			endint, err := strconv.Atoi(limits[2])
			checkY(err)
			endint = minI(endint, maxint)
			query = query + sqlLimit(1+endint-startint,startint)
			dumpRange(w, cred, db, t, o, d, startint, endint, maxint, query)
		} else {
			shipMessage(w, cred, db, "Can't convert to number or range: " + n)
		}
	} else {
		if len(wclauses)>0 {
			dumpWhere(w, cred, db, t, o, d, query, whereQ)
		} else {
			dumpRows(w, cred, db, t, o, d, query)
		}
    }
}


func dumpRows(w http.ResponseWriter, cred Access, db string, t string, o string, d string, query string) {

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)

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
		q.Set("o",o)
		q.Set("d",d)
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

	q.Del("n")
	limitstring := "1-" + strconv.Itoa(rownum-1)
	linkleft := escape("<",q.Encode())
	linkright := escape(">",q.Encode())
	menu := []Entry{}
	menu = append(menu, escape("?",linkselect))
	menu = append(menu, escape("+",linkinsert))
	menu = append(menu, escape("~",linkupdate))
	menu = append(menu, escape("-",linkdeleteF))
	menu = append(menu, escape("i",linkinfo))
	tableOutRows(w, cred, db, t, primary, o, d, limitstring, linkleft, linkright, head, records, menu, "", url.Values{})
}

func dumpWhere(w http.ResponseWriter, cred Access, db string, t string, o string, d string, query string, q url.Values) {

	q.Add("db", db)
	q.Add("t", t)
	q.Del("k")
	q.Del("v")

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
		q.Set("o",o)
		q.Set("d",d)
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
	q.Set("action", "DELETE")
	linkdelete := q.Encode()
	q.Set("action", "INFO")
	linkinfo := q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, escape("?",linkselect))
	menu = append(menu, escape("+",linkinsert))
	menu = append(menu, escape("~",linkupdate))
	menu = append(menu, escape("-",linkdelete))
	menu = append(menu, escape("i",linkinfo))
	wherestring := WhereSelect2Pretty(q, getColumnInfo(cred, db, t))
	tableOutRows(w, cred, db, t, primary, o, d, "", Entry{}, Entry{}, head, records, menu, wherestring, q)
}

func dumpRange(w http.ResponseWriter, cred Access, db string, t string, o string, d string, start int, end int, max int, query string) {

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
