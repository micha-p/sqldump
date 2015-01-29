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
	Escape   string
	Select   string
	Insert   string
	Info     string
	All      string
	X        string
	Left     string
	Right    string
}

func tableOut(w http.ResponseWriter, r *http.Request, records [][]string) {

	u, _, h, p := getCredentials(r)
	db := r.URL.Query().Get("db")
	t := r.URL.Query().Get("t")
	x := r.URL.Query().Get("x")
	var linkleft string
	var linkright string
	var linkall string
	var linkescape string
	var linkselect string
	var linkinsert string
	var linkshow string

	if t == "" {
	} else if x == "" {
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
	} else {
		xint, err := strconv.Atoi(x)
		checkY(err)
		xmax, err := strconv.Atoi(getCount(r, db, t))
		left := strconv.Itoa(maxI(xint-1, 1))
		right := strconv.Itoa(minI(xint+1, xmax))

		q := r.URL.Query()
		q.Set("x", left)
		linkleft = q.Encode()
		q.Set("x", right)
		linkright = q.Encode()
		q.Del("x")
		linkall = q.Encode()
	}

	c := Context{User: u,
		Host:     h,
		Port:     p,
		Database: db,
		Table:    t,
		Records:  records,
		Escape:   href("?"+linkescape, "."),
		Select:   href("?"+linkselect, "/"),
		Insert:   href("?"+linkinsert, "+"),
		Info:     href("?"+linkshow, "?"),
		All:      href("?"+linkall, "."),
		X:        x,
		Left:     href("?"+linkleft, "<"),
		Right:    href("?"+linkright, ">"),
	}

	err := templateTable.Execute(w, c)
	checkY(err)
}
