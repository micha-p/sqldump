package main

/*
<form  action="/login">
   <label for="user">User</label><input type="text"     id="user" name="user"><br>
   <label for="pass">Pass</label><input type="password" id="pass" name="pass"><br>
   <label for="host">Host</label><input type="text"     id="host" name="host" value="localhost"><br>
   <label for="port">Port</label><input type="text"     id="port" name="port" value="3306"><br>
   <button type="submit">Query</button>
</form>
*/

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type FContext struct {
	Action   string
	Selector string
	Button   string
	Database string
	Table    string
	Back     string
	Columns  []string
}

func shipForm(w http.ResponseWriter, r *http.Request, db string, t string, action string, button string, selector string) {

	cols := getCols(r, db, t)
	q := r.URL.Query()
	q.Del("action")
	linkback := q.Encode()

	c := FContext{
		Action:   action,
		Selector: selector,
		Button:   button,
		Database: db,
		Table:    t,
		Back:     linkback,
		Columns:  cols,
	}

	err := templateFormFields.Execute(w, c)
	checkY(err)
}

func actionSubset(w http.ResponseWriter, r *http.Request, database string, table string) {
	shipForm(w, r, database, table, "query", "Query", "true")
}

func actionAdd(w http.ResponseWriter, r *http.Request, database string, table string) {
	shipForm(w, r, database, table, "insert", "Insert", "")
}

func actionQuery(w http.ResponseWriter, r *http.Request) {

	db := r.FormValue("db")
	t := r.FormValue("t")
	cols := getCols(r, db, t)
	v := url.Values{}
	v.Set("db", db)
	v.Set("t", t)
	linkback := "?" + v.Encode()

	var tests []string
	for _, col := range cols {
		val := r.FormValue(col + "C")
		if val != "" {
			comparator := r.FormValue(col + "O")
			if comparator == "" {
				comparator = "="
			}
			tests = append(tests, col+comparator+"\""+val+"\"")
		}
	}

	if len(tests) > 0 {
		// Imploding within templates is severly missing!
		query := "SELECT * FROM " + t + " WHERE " + strings.Join(tests, " && ")
		fmt.Println(query)
		dumpRows(w, r, db, t, linkback, query)
	}
}

func actionInsert(w http.ResponseWriter, r *http.Request) {

	db := r.FormValue("db")
	t := r.FormValue("t")
	cols := getCols(r, db, t)

	// Searching for cols within formValues
	var assignments []string
	for _, col := range cols {
		val := r.FormValue(col + "C")
		if val != "" {
			assignments = append(assignments, col+"=\""+val+"\"")
		}
	}

	if len(assignments) > 0 {
		// Imploding within templates is severly missing!
		stmt := "INSERT INTO " + t + " SET " + strings.Join(assignments, ",")
		user, pw, h, p := getCredentials(r)
		conn, err := sql.Open("mysql", dsn(user, pw, h, p, db))
		checkY(err)
		defer conn.Close()

		statement, err := conn.Prepare(stmt)
		checkY(err)
		_, err = statement.Exec()
		checkY(err)

		http.Redirect(w, r, r.URL.Host+"?db="+db+"&t="+t, 302)
	}
}

/*
+-------+-------------+------+-----+---------+-------+
| Field | Type        | Null | Key | Default | Extra |
+-------+-------------+------+-----+---------+-------+
| title | varchar(64) | YES  |     | NULL    |       |
| start | date        | YES  |     | NULL    |       |
+-------+-------------+------+-----+---------+-------+
*/

func actionShow(w http.ResponseWriter, r *http.Request, db string, t string, back string) {

	rows := getRows(r, db, "show columns from "+template.HTMLEscapeString(t))
	defer rows.Close()

	trail := []Entry{}
	trail = append(trail, Entry{"/", "root"})

	q := url.Values{}
	q.Add("db", db)
	trail = append(trail, Entry{Link: "/?" + q.Encode(), Label: db})
	q.Add("t", t)
	trail = append(trail, Entry{Link: "/?" + q.Encode(), Label: t})

	menu := []Entry{}
	q.Set("action", "add")
	linkinsert := "/?" + q.Encode()
	menu = append(menu, Entry{linkinsert, "+"})
	menu = append(menu, Entry{"/logout", "Q"})

	records := [][]string{}
	head := []string{"Field", "Type", "Null", "Key", "Default", "Extra"}

	var n int = 1
	for rows.Next() {
		var f, t, u, k, e string
		var d []byte // or use http://golang.org/pkg/database/sql/#NullString
		err := rows.Scan(&f, &t, &u, &k, &d, &e)
		checkY(err)
		records = append(records, []string{strconv.Itoa(n), f, t, u, k, string(d), e})
	}
	tableOut(w, r, back, head, records, trail, menu)
}
