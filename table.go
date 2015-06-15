package main

import (
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
	Left     Entry
	Right    Entry
	Trail    []Entry
	Menu     []Entry
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

func makeTrail(host string, db string, t string, primary string, o string, d string, k string, w string, wq url.Values) []Entry {

	q := url.Values{}

	trail := []Entry{Entry{host, "", ""}}

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
	if o != "" {
		wq.Add("o", o)
		if o == primary {
			if d != "" {
				wq.Add("d", d)
				trail = append(trail, escape(makeArrow(o, primary, d), wq.Encode()))
			} else {
				trail = append(trail, escape(makeArrow(o, primary, d), wq.Encode()))
			}
		} else {
			if d != "" {
				wq.Add("d", d)
				trail = append(trail, escape(makeArrow(o, primary, d), wq.Encode()))
			} else {
				trail = append(trail, escape(makeArrow(o, primary, d), wq.Encode()))
			}
		}
	} else if k != "" {
		q.Add("k", k)
		trail = append(trail, escape(k, q.Encode()))
	}
	return trail
}

func makeArrow(title string, primary string, d string) string{
	if title == primary {
		if d == "" {
			return title + "⇑"
		} else {
			return title + "⇓"
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
	head = append(head, escape("#",home.Encode()))

	for _, title := range columns {
		if o == title {
			q.Set("o", title)
			if primary == title {
				if d == "" {
					q.Set("d", "1")
					head = append(head, escape(makeArrow(title, primary, d),q.Encode()))
				} else {
					q.Del("d")
					head = append(head, escape(makeArrow(title, primary, d),q.Encode()))
				}
			} else {
				if d == "" {
					q.Set("d", "1")
					head = append(head, escape(makeArrow(title, primary, d),q.Encode()))
				} else {
					q.Del("d")
					head = append(head, escape(makeArrow(title, primary, d),q.Encode()))
				}
			}
		} else {
			q.Set("o", title)
			q.Del("d")
			head = append(head, escape(title, q.Encode()))
		}
	}
	return head
}

func tableOutSimple(w http.ResponseWriter, cred Access, db string, t string, head []Entry, records [][]Entry, menu []Entry) {

	c := Context{
		User:     cred.User,
		Host:     cred.Host,
		Port:     cred.Port,
		CSS:      CSS_FILE,
		Database: db,
		Table:    t,
		Order:    "",
		Desc:     "",
		Records:  records,
		Head:     head,
		Back:     makeBack(cred.Host, db, t, "", "", ""),
		Counter:  "",
		Left:     Entry{},
		Right:    Entry{},
		Trail:    makeTrail(cred.Host, db, t, "", "", "", "", "", url.Values{}),
		Menu:     menu,
	}
	if DEBUGFLAG {
		initTemplate()
	}
	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutRows(w http.ResponseWriter, cred Access, db string, t string, primary string, o string, d string,
	n string, linkleft Entry, linkright Entry,
	head []Entry, records [][]Entry, menu []Entry, where string, whereQ url.Values) {

	initTemplate()
	c := Context{
		User:     cred.User,
		Host:     cred.Host,
		Port:     cred.Port,
		CSS:      CSS_FILE,
		Database: db,
		Table:    t,
		Order:    o,
		Desc:     d,
		Records:  records,
		Head:     head,
		Back:     makeBack(cred.Host, db, t, "", "", ""),
		Counter:  n,
		Left:     linkleft,
		Right:    linkright,
		Trail:    makeTrail(cred.Host, db, t, primary, o, d, "", where, whereQ),
		Menu:     menu,
	}

	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutFields(w http.ResponseWriter, cred Access, 
	db string, t string, primary string ,o string, d string, k string, n string, 
	linkleft Entry, linkright Entry, head []Entry, records [][]Entry, menu []Entry) {

	initTemplate()

	c := Context{
		User:     cred.User,
		Host:     cred.Host,
		Port:     cred.Port,
		CSS:      CSS_FILE,
		Database: db,
		Table:    t,
		Order:    o,
		Desc:     d,
		Records:  records,
		Head:     head,
		Back:     makeBack(cred.Host, db, t, "", "", ""),
		Counter:  n,
		Left:     linkleft,
		Right:    linkright,
		Trail:    makeTrail(cred.Host, db, t, primary, o, d, k, "", url.Values{}),
		Menu:     menu,
	}

	err := templateTable.Execute(w, c)
	checkY(err)
}
