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
	Left     string
	Right    string
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

func makeTrail(host string, db string, t string, o string, d string, k string, w string, wq url.Values) []Entry {

    q := url.Values{}

	trail := []Entry{Entry{host, ""}}

	if db != "" {
		q.Add("db", db)
		trail = append(trail, Entry{Link: q.Encode(), Text: db})
	}
	if t != "" {
		q.Add("t", t)
		trail = append(trail, Entry{Link: q.Encode(), Text: t})
	}

	wq.Del("o")
	wq.Del("d")
	if w != "" {
		trail = append(trail, Entry{Link: wq.Encode(), Text: w})
	}
	if o != "" {
		wq.Add("o", o)
		if d != "" {
			trail = append(trail, Entry{Link: wq.Encode(), Text: o + "&darr;"})
		} else {
			wq.Add("d", "1")
			trail = append(trail, Entry{Link: wq.Encode(), Text: o + "&uarr;"})
		}
	} else if k != "" {
		q.Add("k", k)
		trail = append(trail, Entry{Link: q.Encode(), Text: k + " (ID)"})
	}
	return trail
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
		Left:     "",
		Right:    "",
		Trail:    makeTrail(cred.Host, db, t, "", "", "", "",url.Values{}),
		Menu:     menu,
	}
	if DEBUGFLAG {
		initTemplate()
	}
	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutRows(w http.ResponseWriter, cred Access, db string, t string, o string, d string, 
	              n string, linkleft string, linkright string, 
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
		Trail:    makeTrail(cred.Host, db, t, o, d, "", where,whereQ),
		Menu:     menu,
	}

	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutFields(w http.ResponseWriter, cred Access, db string, t string, o string, d string, k string, n string, linkleft string, linkright string, head []Entry, records [][]Entry, menu []Entry) {

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
		Trail:    makeTrail(cred.Host, db, t, o, d, k, "",url.Values{}),
		Menu:     menu,
	}

	err := templateTable.Execute(w, c)
	checkY(err)
}
