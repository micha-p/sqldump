package main

import (
	"net/http"
	"fmt"
	"net/url"
)

/*
 * <table>
 * <tr> <th>head 1</th> <th>head 2</th> </tr>
 * <tr> <td>data 1</td> <td>data 2</td> </tr>
 * <tr> <td>data 3</td> <td>data 4</td> </tr>
 * </table>
 */

type Entry struct {
	Link  string
	Label string
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
	Head     []string
	Records  [][]string
	Counter  string
	Left     string
	Right    string
	Trail    []Entry
	Menu     []Entry
}

func makeTrail(host string, db string, t string, o string, d string, k string) []Entry {
	q := url.Values{}
	trail := []Entry{Entry{"/", host}}

	if db != "" {
		q.Add("db", db)
		trail = append(trail, Entry{Link: "?" + q.Encode(), Label: db})
	}
	if t != "" {
		q.Add("t", t)
		trail = append(trail, Entry{Link: "?" + q.Encode(), Label: t})
	}
	if k != "" {
		q.Add("k", k)
		trail = append(trail, Entry{Link: "/?" + q.Encode(), Label: k + " (ID)"})
	} else if o != "" {
		q.Add("o", o)
        if d != "" {
			q.Add("d", d)
			trail = append(trail, Entry{Link: "/?" + q.Encode(), Label: o + "&darr;"})
		} else {
			trail = append(trail, Entry{Link: "/?" + q.Encode(), Label: o + "&uarr;"})
		}
	}
	return trail
}

func shipError(w http.ResponseWriter, cred Access, db string, t string, back string, query string, errormessage error) {

	c := Context{
		User:     cred.User,
		Host:     cred.Host,
		Port:     cred.Port,
		CSS:      CSS_FILE,
		Database: db,
		Table:    t,
		Order:    "",
		Desc:     "",
		Records:  [][]string{},
		Head:     []string{},
		Back:     back,
		Counter:  "",
		Left:     query,
		Right:    fmt.Sprintln(errormessage),
		Trail:    makeTrail(cred.Host,db,t,"","",""),
		Menu:     []Entry{}, 
	}
	if DEBUGFLAG {
		initTemplate()
	}
	err := templateError.Execute(w, c)
	checkY(err)
}

func tableOutSimple(w http.ResponseWriter, cred Access, db string, t string, back string, head []string, records [][]string, menu []Entry) {

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
		Back:     back,
		Counter:  "",
		Left:     "",
		Right:    "",
		Trail:    makeTrail(cred.Host,db,t,"","",""),
		Menu:     menu,
	}
	if DEBUGFLAG {
		initTemplate()
	}
	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutRows(w http.ResponseWriter, cred Access, db string, t string, o string, d string, n string, linkleft string, linkright string, back string, head []string, records [][]string, menu []Entry) {

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
		Back:     back,
		Counter:  n,
		Left:     linkleft,
		Right:    linkright,
		Trail:    makeTrail(cred.Host,db,t,o,d,""),
		Menu:     menu,
	}

	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutFields(w http.ResponseWriter, cred Access, db string, t string, o string, d string, k string, n string, linkleft string, linkright string, back string, head []string, records [][]string, menu []Entry) {

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
		Back:     back,
		Counter:  n,
		Left:     linkleft,
		Right:    linkright,
		Trail:    makeTrail(cred.Host,db,t,o,d,k),
		Menu:     menu,
	}

	err := templateTable.Execute(w, c)
	checkY(err)
}
