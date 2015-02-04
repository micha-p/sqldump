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
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"net/http"
	"strings"
	"strconv"
)

type FContext struct {
	Action   string
	Button   string
	Database string
	Table    string
	Back     string
	Columns  [] string
}


func shipFullForm(w http.ResponseWriter, r *http.Request, db string, t string, action string, button string) {

	rows := getRows(r, db, "select * from "+template.HTMLEscapeString(t))
	defer rows.Close()

	cols, err := rows.Columns()
	checkY(err)

	q := r.URL.Query()
	q.Del("action")
	linkback := q.Encode()


	c := FContext{
		Action:   action,
		Button:   button,
		Database: db,
		Table:    t,
		Back:     linkback,
		Columns:  cols,
	}
	
	err = templateFormFields.Execute(w, c)
	checkY(err)
}

func actionSelect(w http.ResponseWriter, r *http.Request, database string, table string) {
	shipFullForm(w, r, database, table, "select", "Select")
}

func actionInsert(w http.ResponseWriter, r *http.Request, database string, table string) {
	shipFullForm(w, r, database, table, "insert", "Insert")
}

func insertHandler(w http.ResponseWriter, r *http.Request) {
	db := r.FormValue("db")
	t := r.FormValue("t")
	rows := getRows(r, db, "select * from "+template.HTMLEscapeString(t))
	defer rows.Close()

	cols, err := rows.Columns()
	checkY(err)

	// Imploding within templates is severly missing!
	var assignments []string
	for _, col := range cols {
		val := r.FormValue("C" + col)
		if val != "" {
			assignments = append(assignments, "  "+col+"= \""+val+"\"")
		}
	}

	if len(assignments) > 0 {

		stmt := "INSERT INTO " + t + " SET" + strings.Join(assignments, ",")

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

func actionShow(w http.ResponseWriter, r *http.Request, database string, table string, back string) {

	rows := getRows(r, database, "show columns from "+template.HTMLEscapeString(table))
	defer rows.Close()

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
	tableOut(w, r, back, head, records)
}
