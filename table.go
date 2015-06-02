package main

import (
	"net/http"
	"fmt"
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

func shipError(w http.ResponseWriter, cred Access, db string, t string, back string, trail []Entry, query string, errormessage error) {

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
		Trail:    trail, // if trail is missing, menu is shown at the right side of the headline
		Menu:     []Entry{},  // always used. location dependent of presence of trail
	}
	if DEBUGFLAG {
		initTemplate()
	}
	err := templateError.Execute(w, c)
	checkY(err)
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
		Desc:     "",
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

func tableOutRows(w http.ResponseWriter, cred Access, db string, t string, o string, d string, n string, linkleft string, linkright string, back string, head []string, records [][]string, trail []Entry, menu []Entry) {

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
		Trail:    trail,
		Menu:     menu,
	}

	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutFields(w http.ResponseWriter, cred Access, db string, t string, o string, d string, n string, linkleft string, linkright string, back string, head []string, records [][]string, trail []Entry, menu []Entry) {

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
		Trail:    trail,
		Menu:     menu,
	}

	err := templateTable.Execute(w, c)
	checkY(err)
}
