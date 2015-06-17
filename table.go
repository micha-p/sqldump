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
	Label  	 string
	Left     Entry
	Right    Entry
	Trail    []Entry
	Menu     []Entry
	Message  string
	Rows     string
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

func makeTrail(host string, db string, t string, w string, wq url.Values) []Entry {

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
	wq.Del("o")
	wq.Del("d")
	if w != "" {
		trail = append(trail, escape(w, wq.Encode()))
	}
	return trail
}

func makeArrow(title string, primary string, d string) string {
	if primary !="" {
		if d == "" {
			return primary + "↑" // "⇑"
		} else {
			return primary + "↓" // "⇓"
		}
	} else {
		if d == "" {
			return title + "↑"
		} else {
			return title + "↓"
		}
	}
}

func createHead(db string, t string, o string, d string, n string, primary string, columns []string, q url.Values) []Entry {
	head := []Entry{}
	home := url.Values{}
	home.Add("db", db)
	home.Add("t", t)
	head = append(head, escape("#", home.Encode()))

	for _, title := range columns {
		if o == title {
			q.Set("o", title)
			if primary == title {
				if d == "" {
					q.Set("d", "1")
					head = append(head, escape(makeArrow("", primary + " (ID)", d), q.Encode()))
				} else {
					q.Del("d")
					head = append(head, escape(makeArrow("", primary + " (ID)", d), q.Encode()))
				}
			} else {
				if d == "" {
					q.Set("d", "1")
					head = append(head, escape(makeArrow(title, "", d), q.Encode()))
				} else {
					q.Del("d")
					head = append(head, escape(makeArrow(title, "", d), q.Encode()))
				}
			}
		} else {
			q.Set("o", title)
			q.Del("d")
			if primary == title {
				head = append(head, escape(title + " (ID)", q.Encode()))
			} else {
				head = append(head, escape(title, q.Encode()))
			}
		}
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
		Label:	  "",
		Left:     Entry{},
		Right:    Entry{},
		Trail:    makeTrail(host, db, t, "", url.Values{}),
		Menu:     menu,
		Message:  "",
		Rows:	  "",
	}
	if DEBUGFLAG {
		initTemplate()
	}
	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutRows(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, primary string, o string, d string,
	n string, counterLabel string, linkleft Entry, linkright Entry,
	head []Entry, records [][]Entry, menu []Entry, msg string, rows string, where string, whereQ url.Values) {

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
		Counter:  n,
		Label:	  counterLabel,
		Left:     linkleft,
		Right:    linkright,
		Trail:    makeTrail(host, db, t, where, whereQ),
		Menu:     menu,
		Message:  msg,
		Rows:	  rows,
	}

	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutFields(w http.ResponseWriter, conn *sql.DB, host string,
	db string, t string, primary string, o string, d string, k string, n string, counterLabel string,
	linkleft Entry, linkright Entry, head []Entry, records [][]Entry, menu []Entry) {

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
		Counter:  n,
		Label:	  counterLabel,
		Left:     linkleft,
		Right:    linkright,
		Trail:    makeTrail(host, db, t, "", url.Values{}),
		Menu:     menu,
		Message:  "",
		Rows:	  "",
	}

	err := templateTable.Execute(w, c)
	checkY(err)
}
