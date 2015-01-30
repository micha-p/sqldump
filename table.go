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

type Context struct {
	User     string
	Host     string
	Port     string
	Database string
	Table    string
	Records  [][]string
	Back     string
	Select   string
	Insert   string
	Left     string
	X        string
	Right    string
	Info     string
}

func tableOut(w http.ResponseWriter, r *http.Request, records [][]string) {

	u, _, h, p := getCredentials(r)
	db := r.URL.Query().Get("db")
	t := r.URL.Query().Get("t")
	var linkescape string
	var linkselect string
	var linkinsert string
	var linkshow string

	if t != "" {
		q := r.URL.Query()
		q.Add("action", "insert")
		linkinsert = q.Encode()
		q.Del("action")
		q.Add("action", "select")
		linkselect = q.Encode()
		q.Del("action")
		q.Add("action", "show")
		linkshow = q.Encode()
		q.Del("action")
		q.Del("t")
		linkescape = q.Encode()
	}

	c := Context{
		User:     u,
		Host:     h,
		Port:     p,
		Database: db,
		Table:    t,
		Records:  records,
		Back:     href("?"+linkescape, "."),
		Select:   href("?"+linkselect, "/"),
		Insert:   href("?"+linkinsert, "+"),
		Left:     "",
		X:        "",
		Right:    "",
		Info:     href("?"+linkshow, "?"),
	}
	err := templateTable.Execute(w, c)
	checkY(err)
}

func tableOutFields(w http.ResponseWriter, r *http.Request, records [][]string) {

	u, _, h, p := getCredentials(r)
	db := r.URL.Query().Get("db")
	t := r.URL.Query().Get("t")
	n := r.URL.Query().Get("n")

	var linkleft string
	var linkright string
	var linkall string
	var linkshow string

	nint, err := strconv.Atoi(n)
	nmax, err := strconv.Atoi(getCount(r, db, t))
	left := strconv.Itoa(maxI(nint-1, 1))
	right := strconv.Itoa(minI(nint+1, nmax))

	q := r.URL.Query()
	q.Set("n", left)
	linkleft = q.Encode()
	q.Set("n", right)
	linkright = q.Encode()
	q.Del("n")
	linkall = q.Encode()

	c := Context{User: u,
		Host:     h,
		Port:     p,
		Database: db,
		Table:    t,
		Records:  records,
		Back:     href("?"+linkall, "."),
		Select:   "",
		Insert:   "",
		Left:     href("?"+linkleft, "<"),
		X:        n,
		Right:    href("?"+linkright, ">"),
		Info:     href("?"+linkshow, "?"),
	}

	err = templateTableFields.Execute(w, c)
	checkY(err)
}
