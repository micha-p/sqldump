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

func shipForm(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, action string, button string, selector string) {

	cols := getCols(cred, db, t)
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

func actionSubset(w http.ResponseWriter, r *http.Request, cred Access, database string, table string) {
	shipForm(w, r, cred, database, table, "query", "Query", "true")
}

func actionAdd(w http.ResponseWriter, r *http.Request, cred Access, database string, table string) {
	shipForm(w, r, cred, database, table, "insert", "Insert", "")
}

func actionQuery(w http.ResponseWriter, r *http.Request, cred Access) {

	db := r.FormValue("db")
	t := r.FormValue("t")
	cols := getCols(cred, db, t)
	v := url.Values{}
	v.Set("db", db)
	v.Set("t", t)
	linkback := "?" + v.Encode()

	trail := []Entry{}
	trail = append(trail, Entry{"/", cred.Host})

	q := url.Values{}
	q.Add("db", db)
	trail = append(trail, Entry{Link: "/?" + q.Encode(), Label: db})
	q.Add("t", t)
	trail = append(trail, Entry{Link: "/?" + q.Encode(), Label: t})

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
		trail = append(trail, Entry{Link: "/?" + r.URL.RawQuery, Label: strings.Join(tests, " ")})
		dumpRows(w, r, cred, trail, db, t, linkback, query)
	}
}

func actionInsert(w http.ResponseWriter, r *http.Request, cred Access) {

	db := r.FormValue("db")
	t := r.FormValue("t")
	cols := getCols(cred, db, t)

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
		conn := getConnection(cred, db)
		defer conn.Close()

		statement, err := conn.Prepare(stmt)
		checkY(err)
		_, err = statement.Exec()
		checkY(err)
		fmt.Println(stmt)
		http.Redirect(w, r, r.URL.Host+"?db="+db+"&t="+t, 302)
	}
}

/*
 show columns from posts;
+-------+-------------+------+-----+---------+-------+
| Field | Type        | Null | Key | Default | Extra |
+-------+-------------+------+-----+---------+-------+
| title | varchar(64) | YES  |     | NULL    |       |
| start | date        | YES  |     | NULL    |       |
+-------+-------------+------+-----+---------+-------+
*/

func actionShow(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, back string) {

	rows := getRows(cred, db, "show columns from "+template.HTMLEscapeString(t))
	defer rows.Close()

	trail := []Entry{}
	trail = append(trail, Entry{"/", cred.Host})

	q := url.Values{}
	q.Add("db", db)
	trail = append(trail, Entry{Link: "/?" + q.Encode(), Label: db})
	q.Add("t", t)
	trail = append(trail, Entry{Link: "/?" + q.Encode(), Label: t})

	menu := []Entry{}
	q.Set("action", "add")
	linkinsert := "/?" + q.Encode()
	menu = append(menu, Entry{linkinsert, "+"})

	records := [][]string{}
	head := []string{"Field", "Type", "Null", "Key", "Default", "Extra"}

	var n int = 1
	for rows.Next() {
		var f, t, u, k, e string
		var d []byte // or use http://golang.org/pkg/database/sql/#NullString
		err := rows.Scan(&f, &t, &u, &k, &d, &e)
		checkY(err)
		records = append(records, []string{strconv.Itoa(n), f, t, u, k, string(d), e})
		n = n + 1
	}
	tableOut(w, r, cred, back, head, records, trail, menu)
}
