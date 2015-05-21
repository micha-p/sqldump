package main

import (
	"net/http"
	"strconv"
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
	CSS		 string
	Database string
	Table    string
	Back     string
	Head     []string
	Records  [][]string
	Select   string
	Insert   string
	Left	 string
	Counter  string
	This  	 string
	Right    string
	Info     string
	Trail    []Entry
	Menu     []Entry
}

func tableOut(w http.ResponseWriter, r *http.Request, cred Access, back string, head []string, records [][]string, trail []Entry, menu []Entry) {

	initTemplate()
	db := r.URL.Query().Get("db")
	t := r.URL.Query().Get("t")
	var linkwhere string
	var linkinsert string
	var linkshow string

	if t != "" {
		q := r.URL.Query()
		q.Add("action", "insert")
		linkinsert = q.Encode()
		q.Del("action")
		q.Add("action", "where")
		linkwhere = q.Encode()
		q.Del("action")
		q.Add("action", "show")
		linkshow = q.Encode()
		q.Del("action")
	}

	c := Context{
		User:     cred.User,
		Host:     cred.Host,
		Port:     cred.Port,
		CSS:      CSS_FILE,
		Database: db,
		Table:    t,
		Records:  records,
		Head:     head,
		Back:     back,
		Select:   href("?"+linkwhere, "/"),
		Insert:   href("?"+linkinsert, "+"),
		Left:       "",
		Counter:  "",
		This:     "",
		Right:     "",
		Info:     href("?"+linkshow, "?"),
		Trail:    trail, // if trail is missing, menu is shown at the right side of the headline
		Menu:     menu,  // always used. location dependent of presence of trail
	}
	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutFields(w http.ResponseWriter, r *http.Request, cred Access, back string, head []string, records [][]string, trail []Entry, menu []Entry) {

	initTemplate()
	db := r.URL.Query().Get("db")
	t := r.URL.Query().Get("t")
	n := r.URL.Query().Get("n")

	var linkleft string
	var linkthis string
	var linkright string
	var linkshow string

	nint, err := strconv.Atoi(n)
	nmax, err := strconv.Atoi(getCount(cred, db, t))
	left := strconv.Itoa(maxI(nint-1, 1))
	right := strconv.Itoa(minI(nint+1, nmax))

	q := r.URL.Query()
	q.Set("n", left)
	linkleft = "?" + q.Encode()
	q.Set("n", n)
	linkthis = "?" + q.Encode()
	q.Set("n", right)
	linkright = "?" + q.Encode()

	c := Context{
		User:     cred.User,
		Host:     cred.Host,
		Port:     cred.Port,
		CSS:      CSS_FILE,
		Database: db,
		Table:    t,
		Records:  records,
		Head:     head,
		Back:     back,
		Select:   "",
		Insert:   "",
		Left:     linkleft,
		Counter:  n,
		This: 	  linkthis,
		Right:    linkright,
		Info:     href("?"+linkshow, "?"),
		Trail:    trail,
		Menu:     menu,
	}

	err = templateTable.Execute(w, c)
	checkY(err)
}
