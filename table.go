package main

import (
	"database/sql"
	"net/http"
	"net/url"
)

/*
 * <table>
 * <tr> <th>head 1</th> <th>head 2</th> </tr>
 * <tr> <td>data 1</td> <td>data 2</td> </tr>
 * <tr> <td>data 3</td> <td>data 4</td> </tr>
 * </table>
 */

// all output has to be supplied as struct to respective template
type Entry struct {
	Text string
	Link string
	Null string
}

type Message struct {
	Msg      string
	Rows     int64
	Affected int64
	Seconds  float64
}

type TContext struct {
	User     string
	Host     string
	Port     string
	CSS      string
	Database string
	Table    string
	Order    string
	Desc     string
	Back     string
	Head     []Entry
	Records  [][]Entry
	Counter  string
	Label    string
	Hidden   []CContext
	Left     Entry
	Right    Entry
	Trail    []Entry
	Menu     []Entry
	Messages []Message
	Rows     int
	Affected int
	Seconds  float64
}

func makeBack(host string, db string, t string, o string, d string, k string) string {
	q := url.Values{}
	if db != "" {
		if t != "" {
			q.Add("t", t)
		}
		if k != "" {
			q.Add("k", k)
		} else if o != "" {
			q.Add("o", o)
			if d != "" {
				q.Add("d", d)
			}
		}
		return q.Encode()
	} else {
		return "/logout"
	}
}

func makeTrail(host string, db string, t string, o string, d string, whereStack [][]Clause) []Entry {

	q := url.Values{}

	trail := []Entry{escape(host, "logout",""), escape(db, q.Encode())}
	if t != "" {
		q.Add("t", t)
		trail = append(trail, escape(t, q.Encode()))
	}

	q.Set("t", t)
	q.Set("o", o)
	q.Set("d", d)
	for i, whereLevel := range whereStack {
		putWhereStackIntoQuery(q, whereStack[0:i+1])
		trail = append(trail, escape(whereClauses2Pretty(whereLevel), q.Encode()))
	}
	return trail
}

func makeTitleWithArrow(title string, primary string, d string) string {
	if title == primary {
		if d == "" {
			return primary + " (ID)" + "↑" // "⇑"
		} else {
			return primary + " (ID)" + "↓" // "⇓"
		}
	} else {
		if d == "" {
			return title + "↑"
		} else {
			return title + "↓"
		}
	}
}

func makeTitleEntry(path string, q url.Values, column string, primary string, o string, d string) Entry {
	var r Entry
	if o == column {
		label := makeTitleWithArrow(column, primary, d)
		q.Set("o", o)
		if d == "" {
			q.Set("d", "1")
			r = escape(label, path, q.Encode())
			q.Del("d")
		} else {
			q.Del("d")
			r = escape(label, path, q.Encode())
			q.Set("d", d)
		}
		q.Set("o", o)
	} else {
		q.Set("o", column)
		q.Del("d")
		if primary == column {
			r = escape(column+" (ID)", path, q.Encode())
		} else {
			r = escape(column, path, q.Encode())
		}
	}
	q.Set("o", o)
	q.Set("d", d)
	return r
}

func createHead(t string, o string, d string, n string, primary string, columns []string, path string, q url.Values) []Entry {
	head := []Entry{}
	home := url.Values{}
	if t !="" {
		home.Set("t", t)
		q.Set("t", t)
	}
	head = append(head, escape("#", path, home.Encode()))

	q.Set("t", t)
	for _, title := range columns {
		head = append(head, makeTitleEntry(path, q, title, primary, o, d))
	}
	return head
}

func tableOutSimple(w http.ResponseWriter, conn *sql.DB, t string, head []Entry, records [][]Entry, menu []Entry) {

	host,db := getHostDB(getDSN(conn))
	c := TContext{
		User:     "",
		Host:     host,
		Port:     "",
		CSS:      CSS_FILE,
		Database: db,
		Table:    t,
		Order:    "",
		Desc:     "",
		Records:  records,
		Head:     head,
		Back:     makeBack(host, db, t, "", "", ""),
		Counter:  "",
		Label:    "",
		Hidden:   []CContext{},
		Left:     Entry{},
		Right:    Entry{},
		Trail:    makeTrail(host, db, t, "", "", [][]Clause{}),
		Menu:     menu,
		Messages: []Message{},
	}
	if DEBUGFLAG {
		initTemplate()
	}
	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutRows(w http.ResponseWriter, conn *sql.DB, t string, primary string, o string, d string,
	n string, counterLabel string, linkleft Entry, linkright Entry,
	head []Entry, records [][]Entry, menu []Entry, messageStack []Message, whereStack [][]Clause) {

	var msgs []Message
	if !QUIETFLAG {
		msgs = messageStack
	}

	host,db := getHostDB(getDSN(conn))
	c := TContext{
		User:     "",
		Host:     host,
		Port:     "",
		CSS:      CSS_FILE,
		Database: db,
		Table:    t,
		Order:    o,
		Desc:     d,
		Records:  records,
		Head:     head,
		Back:     makeBack(host, db, t, "", "", ""),
		Counter:  n, // TODO counter needs hidden cols for where clauses
		Label:    counterLabel,
		Hidden:   WhereStack2Hidden(whereStack),
		Left:     linkleft,
		Right:    linkright,
		Trail:    makeTrail(host, db, t, o, d, whereStack),
		Menu:     menu,
		Messages: msgs,
	}

	if DEBUGFLAG {
		initTemplate()
	}
	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutFields(w http.ResponseWriter, conn *sql.DB, t string, o string, d string,
	counterContent string, counterLabel string,
	linkleft Entry, linkright Entry,
	head []Entry, records [][]Entry, menu []Entry, whereStack [][]Clause) {

	initTemplate()

	host,db := getHostDB(getDSN(conn))
	c := TContext{
		User:     "",
		Host:     host,
		Port:     "",
		CSS:      CSS_FILE,
		Database: db,
		Table:    t,
		Order:    o,
		Desc:     d,
		Records:  records,
		Head:     head,
		Back:     makeBack(host, db, t, "", "", ""),
		Counter:  counterContent,
		Label:    counterLabel,
		Hidden:   WhereStack2Hidden(whereStack),
		Left:     linkleft,
		Right:    linkright,
		Trail:    makeTrail(host, db, t, "", "", whereStack),
		Menu:     menu,
		Messages: []Message{},
	}

	err := templateTable.Execute(w, c)
	checkY(err)
}
