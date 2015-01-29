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
	"strings"
	"html/template"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
)

const formA = "<form  action=\"/%s\">\n"
const formTempl = "   <label for=\"{{.}}\">{{.}}</label><input type=\"text\" id=\"{{.}}\" name=\"{{.}}\"><br>\n"
const formTemplD = "   <input type=\"hidden\" name=\"db\" value=\"{{.}}\">\n"
const formTemplT = "   <input type=\"hidden\" name=\"t\"  value=\"{{.}}\">\n"
const formO = "<button type=\"submit\">%s</button>\n</form>\n"

func shipFormline(w http.ResponseWriter, s string) {
	t, err := template.New("th").Parse(formTempl)
	checkY(err)
	err = t.Execute(w, s)
	checkY(err)
}

func shipFormlineD(w http.ResponseWriter, s string) {
	t, err := template.New("th").Parse(formTemplD)
	checkY(err)
	err = t.Execute(w, s)
	checkY(err)
}

func shipFormlineT(w http.ResponseWriter, s string) {
	t, err := template.New("th").Parse(formTemplT)
	checkY(err)
	err = t.Execute(w, s)
	checkY(err)
}

func shipForm(w http.ResponseWriter, r *http.Request, database string, table string, action string, button string) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	rows := getRows(r, database, "select * from "+template.HTMLEscapeString(table))
	defer rows.Close()

	cols, err := rows.Columns()
	checkY(err)

	fmt.Fprintf(w, formA, action)
	shipFormlineD(w, database)
	shipFormlineT(w, table)
	for _, col := range cols {
		shipFormline(w, col)
	}
	fmt.Fprintf(w, formO, button)
}

func actionSelect(w http.ResponseWriter, r *http.Request, database string, table string) {
	shipForm(w, r, database, table, "select", "Select")
}

func actionInsert(w http.ResponseWriter, r *http.Request, database string, table string) {
	shipForm(w, r, database, table, "insert", "Insert")
}

// http://jan.newmarch.name/go/template/chapter-template.html
// http://golang.org/pkg/text/template/

const sqlinsert = `insert into {{.}} set {{range pairs}}`

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
		val := r.FormValue(col)
		if val != "" {
			assignments = append (assignments, "  " + col + "= \"" + val + "\"")
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
	
	    http.Redirect(w, r, r.URL.Host + "?db=" + db + "&t=" + t, 302)
	}
}
