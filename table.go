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

type Context struct {
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
		q.Add("db", db)
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

func makeTrail(host string, db string, t string, o string, d string, w string, wq url.Values) []Entry {

	q := url.Values{}

	trail := []Entry{Entry{host, "/", ""}}

	if db != "" {
		q.Add("db", db)
		trail = append(trail, escape(db, q.Encode()))
	}
	if t != "" {
		q.Add("t", t)
		trail = append(trail, escape(t, q.Encode()))
	}

	wq.Set("db", db)
	wq.Set("t", t)
	wq.Set("o", o)
	wq.Set("d", w)
	if w != "" {
		trail = append(trail, escape(w, wq.Encode()))
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

func makeTitleEntry(q url.Values, column string, primary string, o string, d string) Entry {
	var r Entry
	if o == column {
		label := makeTitleWithArrow(column, primary, d)
		q.Set("o", o)
		if d == "" {
			q.Set("d", "1")
			r = escape(label, q.Encode())
			q.Del("d")
		} else {
			q.Del("d")
			r = escape(label, q.Encode())
			q.Set("d", d)
		}
		q.Set("o", o)
	} else {
		q.Set("o", column)
		q.Del("d")
		if primary == column {
			r = escape(column+" (ID)", q.Encode())
		} else {
			r = escape(column, q.Encode())
		}
	}
	q.Set("o", o)
	q.Set("d", d)
	return r
}

func createHead(db string, t string, o string, d string, n string, primary string, columns []string, q url.Values) []Entry {
	head := []Entry{}
	home := url.Values{}
	home.Set("db", db)
	home.Set("t", t)
	head = append(head, escape("#", home.Encode()))

	q.Set("db", db)
	q.Set("t", t)
	for _, title := range columns {
		head = append(head, makeTitleEntry(q, title, primary, o, d))
	}
	return head
}

func tableOutSimple(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, head []Entry, records [][]Entry, menu []Entry) {

	c := Context{
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
		Left:     Entry{},
		Right:    Entry{},
		Trail:    makeTrail(host, db, t, "", "","",url.Values{}),
		Menu:     menu,
		Messages: []Message{},
	}
	if DEBUGFLAG {
		initTemplate()
	}
	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutRows(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, primary string, o string, d string,
	n string, counterLabel string, linkleft Entry, linkright Entry,
	head []Entry, records [][]Entry, menu []Entry, messageStack []Message, whereString string, whereQuery url.Values) {

	var msgs []Message
	if !QUIETFLAG {
		msgs = messageStack
	}

	c := Context{
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
		Counter:  n,
		Label:    counterLabel,
		Left:     linkleft,
		Right:    linkright,
		Trail:	  makeTrail(host, db, t, o,d, whereString, whereQuery),
		Menu:     menu,
		Messages: msgs,
	}

	if DEBUGFLAG {
		initTemplate()
	}
	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutFields(w http.ResponseWriter, conn *sql.DB, host string,
	db string, t string, o string, d string,
	counterContent string, counterLabel string,
	linkleft Entry, linkright Entry,
	head []Entry, records [][]Entry, menu []Entry, whereString string, whereQuery url.Values) {

	initTemplate()

	c := Context{
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
		Left:     linkleft,
		Right:    linkright,
		Trail:    makeTrail(host, db, t, "", "", whereString, whereQuery),
		Menu:     menu,
		Messages: []Message{},
	}

	err := templateTable.Execute(w, c)
	checkY(err)
}
