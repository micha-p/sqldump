package main

import (
	"net/http"
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
	OrderD   string
	Back     string
	Head     []string
	Records  [][]string
	Counter  string
	Left     string
	Right    string
	Trail    []Entry
	Menu     []Entry
}

func tableOut(w http.ResponseWriter, cred Access, db string, t string, back string, head []string, records [][]string, trail []Entry, menu []Entry) {

	c := Context{
		User:     cred.User,
		Host:     cred.Host,
		Port:     cred.Port,
		CSS:      CSS_FILE,
		Database: db,
		Table:    t,
		Order:    "",
		OrderD:   "",
		Records:  records,
		Head:     head,
		Back:     back,
		Counter:  "",
		Left:     "",
		Right:    "",
		Trail:    trail, // if trail is missing, menu is shown at the right side of the headline
		Menu:     menu,  // always used. location dependent of presence of trail
	}
	if DEBUGFLAG {
		initTemplate()
	}
	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutRows(w http.ResponseWriter, cred Access, db string, t string, o string, od string, n string, linkleft string, linkright string, back string, head []string, records [][]string, trail []Entry, menu []Entry) {

	initTemplate()

	c := Context{
		User:     cred.User,
		Host:     cred.Host,
		Port:     cred.Port,
		CSS:      CSS_FILE,
		Database: db,
		Table:    t,
		Order:    o,
		OrderD:   od,
		Records:  records,
		Head:     head,
		Back:     back,
		Counter:  n,
		Left:     linkleft,
		Right:    linkright,
		Trail:    trail,
		Menu:     menu,
	}

	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutFields(w http.ResponseWriter, cred Access, db string, t string, o string, od string, n string, linkleft string, linkright string, back string, head []string, records [][]string, trail []Entry, menu []Entry) {

	initTemplate()

	c := Context{
		User:     cred.User,
		Host:     cred.Host,
		Port:     cred.Port,
		CSS:      CSS_FILE,
		Database: db,
		Table:    t,
		Order:    o,
		OrderD:   od,
		Records:  records,
		Head:     head,
		Back:     back,
		Counter:  n,
		Left:     linkleft,
		Right:    linkright,
		Trail:    trail,
		Menu:     menu,
	}

	err := templateTable.Execute(w, c)
	checkY(err)
}
