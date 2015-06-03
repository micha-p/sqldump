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
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type CContext struct {
	Number    string
	Name      string
	Label     string
	IsNumeric string
	IsString  string
	Value     string
	Readonly  string
}

type FContext struct {
	CSS      string
	Action   string
	Selector string
	Button   string
	Database string
	Table    string
	Key      string
	Value    string
	Back     string
	Columns  []CContext
	Trail    []Entry
}

func shipErrorPage(w http.ResponseWriter, cred Access, db string, t string, cols []CContext) {

	c := FContext{
		CSS:      CSS_FILE,
		Action:   "BACK",
		Selector: "",
		Button:   "Back",
		Database: db,
		Table:    t,
		Back:     makeBack(cred.Host, db, t, "", "", ""),
		Columns:  cols,
		Trail:    makeTrail(cred.Host, db, t, "", "", "", ""),
	}

	if DEBUGFLAG {
		initTemplate()
	}
	err := templateError.Execute(w, c)
	checkY(err)
}

func shipError(w http.ResponseWriter, cred Access, db string, t string, query string, e error) {
	cols := []CContext{CContext{"1", "", "Query", "", "", query, ""}, CContext{"2", "", "Result", "", "", fmt.Sprint(e), ""}}
	shipErrorPage(w, cred, db, t, cols)
}

func shipMessage(w http.ResponseWriter, cred Access, db string, msg string) {
	cols := []CContext{CContext{"1", "", "Message", "", "", msg, ""}}
	shipErrorPage(w, cred, db, "", cols)
}

func shipForm(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, action string, button string, selector string) {

	cols := getColumnInfo(cred, db, t)
	primary := getPrimary(cred, db, t)
	q := r.URL.Query()
	q.Del("action")
	linkback := q.Encode()
	newcols := []CContext{}

	for _, col := range cols {
		label := col.Name
		if label == primary {
			label = label + " (ID)"
		}
		newcols = append(newcols, CContext{col.Number, col.Name, label, col.IsNumeric, col.IsString, "", ""})
	}

	c := FContext{
		CSS:      CSS_FILE,
		Action:   action,
		Selector: selector,
		Button:   button,
		Database: db,
		Table:    t,
		Key:      "",
		Value:    "",
		Back:     linkback,
		Columns:  newcols,
		Trail:    makeTrail(cred.Host, db, t, "", "", "", ""),
	}

	if DEBUGFLAG {
		initTemplate()
	}
	err := templateFormFields.Execute(w, c)
	checkY(err)
}

func shipFormWithValues(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, k string, v string, action string, button string, selector string, query string) {

	cols := getColumnInfo(cred, db, t)
	fieldmap := getFieldMap(w, db, t, cred, query)
	primary := getPrimary(cred, db, t)
	newcols := []CContext{}

	for _, col := range cols {
		label := col.Name
		readonly := ""
		if label == primary {
			label = label + " (ID)"
			readonly = "true"
		}
		newcols = append(newcols, CContext{col.Number, col.Name, label, col.IsNumeric, col.IsString, fieldmap[col.Name], readonly})
	}

	q := r.URL.Query()
	q.Del("action")
	linkback := q.Encode()

	c := FContext{
		CSS:      CSS_FILE,
		Action:   action,
		Selector: selector,
		Button:   button,
		Database: db,
		Table:    t,
		Key:      k,
		Value:    v,
		Back:     linkback,
		Columns:  newcols,
		Trail:    makeTrail(cred.Host, db, t, "", "", "", ""),
	}

	if DEBUGFLAG {
		initTemplate()
	}
	err := templateFormFields.Execute(w, c)
	checkY(err)
}

func actionSubset(w http.ResponseWriter, r *http.Request, cred Access, database string, table string) {
	shipForm(w, r, cred, database, table, "QUERY", "Query", "true")
}

func actionDelete(w http.ResponseWriter, r *http.Request, cred Access, database string, table string) {
	shipForm(w, r, cred, database, table, "DELETEEXEC", "Delete", "true")
}

func actionAdd(w http.ResponseWriter, r *http.Request, cred Access, database string, table string) {
	shipForm(w, r, cred, database, table, "INSERT", "Insert", "")
}

func actionEdit(w http.ResponseWriter, r *http.Request, cred Access, database string, t string, k string, v string) {
	query := "select * from " + t + " where " + k + "=" + v
	shipFormWithValues(w, r, cred, database, t, k, v, "EDITEXEC", "Submit", "", query)
}

func actionQuery(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {

	cols := getCols(cred, db, t)
	var clauses []string
	for _, col := range cols {
		val := sqlprotect(r.FormValue(col + "C"))
		if val != "" {
			comparator := sqlprotect(r.FormValue(col + "O"))
			if comparator == "" {
				clauses = append(clauses, col+sqlFilterNumeric(val))
			} else {
				clauses = append(clauses, col+comparator+"\""+val+"\"")
			}
		}
	}

	if len(clauses) > 0 {
		where := strings.Join(clauses, " && ")
		query := "SELECT * FROM " + t + " WHERE " + where
		dumpRows(w, db, t, "", "", cred, query, where)
	}
}

func actionDeleteExec(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {

	cols := getCols(cred, db, t)
	var clauses []string
	for _, col := range cols {
		val := sqlprotect(r.FormValue(col + "C"))
		if val != "" {
			comparator := sqlprotect(r.FormValue(col + "O"))
			if comparator == "" {
				clauses = append(clauses, col+sqlFilterNumeric(val))
			} else {
				clauses = append(clauses, col+comparator+"\""+val+"\"")
			}
		}
	}
	if len(clauses) > 0 {
		stmt := "DELETE FROM " + t + " WHERE " + strings.Join(clauses, " && ")
		log.Println("[SQL]", stmt)
		conn := getConnection(cred, db)
		defer conn.Close()

		statement, err := conn.Prepare(stmt)
		checkY(err)
		_, err = statement.Exec()
		checkY(err)
		http.Redirect(w, r, r.URL.Host+"?db="+db+"&t="+t, 302)
	}
}

func actionDeleteWhere(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, where string) {
	stmt := "DELETE FROM " + t + " WHERE " + where
	log.Println("[SQL]", stmt)
	conn := getConnection(cred, db)
	defer conn.Close()

	statement, err := conn.Prepare(stmt)
	checkY(err)
	_, err = statement.Exec()
	checkY(err)
	http.Redirect(w, r, r.URL.Host+"?db="+db+"&t="+t, 302)
}

func actionInsert(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {

	cols := getCols(cred, db, t)
	// Searching for cols within formValues
	var clauses []string
	for _, col := range cols {
		val := sqlprotect(r.FormValue(col + "C"))
		if val != "" {
			clauses = append(clauses, col+"=\""+val+"\"")
		}
	}

	if len(clauses) > 0 {
		// Imploding within templates is severly missing!
		stmt := "INSERT INTO " + t + " SET " + strings.Join(clauses, ",")
		log.Println("[SQL]", stmt)
		conn := getConnection(cred, db)
		defer conn.Close()

		statement, err := conn.Prepare(stmt)
		checkY(err)
		_, err = statement.Exec()
		checkY(err)
		http.Redirect(w, r, r.URL.Host+"?db="+db+"&t="+t, 302)
	}
}

func actionEditExec(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, k string, v string) {

	cols := getCols(cred, db, t)
	// Searching for cols within formValues
	var clauses []string
	for _, col := range cols {
		val := sqlprotect(r.FormValue(col + "C"))
		if val != "" && col != k {
			clauses = append(clauses, col+"=\""+val+"\"")
		}
	}

	if len(clauses) > 0 {
		// Imploding within templates is severly missing!
		// TODO stmt := "UPDATE ? SET " + strings.Join(clauses, ",") + " where ? = ?
		stmt := "UPDATE " + t + " SET " + strings.Join(clauses, ",") + " where " + k + "=" + v
		log.Println("[SQL]", stmt)
		conn := getConnection(cred, db)
		defer conn.Close()

		statement, err := conn.Prepare(stmt)
		checkY(err)
		_, err = statement.Exec()
		checkY(err)
		http.Redirect(w, r, r.URL.Host+"?db="+db+"&t="+t, 302)
	}
}

func actionRemove(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, k string, v string) {

	stmt := "DELETE FROM " + t + " WHERE " + k + "=" + v
	log.Println("[SQL]", stmt)
	conn := getConnection(cred, db)
	defer conn.Close()

	statement, err := conn.Prepare(stmt)
	checkY(err)
	_, err = statement.Exec()
	checkY(err)
	http.Redirect(w, r, r.URL.Host+"?db="+db+"&t="+t+"&k"+k, 302)
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

func actionInfo(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {

	rows, err := getRows(cred, db, "show columns from "+t)
	checkY(err)
	defer rows.Close()

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)

	menu := []Entry{}
	q.Set("action", "ADD")
	linkinsert := "/?" + q.Encode()
	menu = append(menu, Entry{linkinsert, "+"})

	records := [][]string{}
	head := []string{"#", "Field", "Type", "Null", "Key", "Default", "Extra"}

	var i int = 1
	for rows.Next() {
		var f, t, n, k, e string
		var d []byte // or use http://golang.org/pkg/database/sql/#NullString
		err := rows.Scan(&f, &t, &n, &k, &d, &e)
		checkY(err)
		records = append(records, []string{strconv.Itoa(i), f, t, n, k, string(d), e})
		i = i + 1
	}
	tableOutSimple(w, cred, db, t, head, records, menu)
}
