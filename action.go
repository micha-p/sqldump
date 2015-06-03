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
	"html"
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

func shipForm(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, k string, v string, action string, button string, selector string, fieldmap map[string]string) {

	cols := getColumnInfo(cred, db, t)
	primary := getPrimary(cred, db, t)
	newcols := []CContext{}

	for _, col := range cols {
		label := col.Name
		readonly := ""
		value := html.EscapeString(fieldmap[col.Name])
		if label == primary {
			label = label + " (ID)"
			readonly = value
		}
		newcols = append(newcols, CContext{col.Number, col.Name, label, col.IsNumeric, col.IsString, value, readonly})
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

func actionSubset(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {
	shipForm(w, r, cred, db, t, "", "", "QUERY", "Query", "true", make(map[string]string) )
}

func actionDelete(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {
	shipForm(w, r, cred, db, t, "","", "DELETEEXEC", "Delete", "true", make(map[string]string) )
}

func actionAdd(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {
	shipForm(w, r, cred, db, t, "","","INSERT", "Insert", "", make(map[string]string) )
}

func actionEdit(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, k string, v string) {
	query := "select * from `" + t + "` where `" + k + "` = ?" 
	conn := getConnection(cred, db)
	defer conn.Close()
	stmt, err := conn.Prepare(query)
	checkY(err)	
	rows, err := stmt.Query(v)
	checkY(err)
	shipForm(w, r, cred, db, t, k, v, "EDITEXEC", "Submit", "", getFieldMap(w, db, t, cred, rows))
}



func collectClauses(r *http.Request, cred Access, db string, t string, set string) []string {

	var clauses []string
	cols := getCols(cred, db, t)
	for _, col := range cols {
		val := sqlProtectString(r.FormValue(col + "C"))
		if val != "" {
			comparator := sqlProtectString(r.FormValue(col + "O"))
			if comparator == "" {
				if set == "" {
					clauses = append(clauses, "`" + col + "`" + sqlProtectNumericComparison(val))
				} else {
					clauses = append(clauses, "`" + col + "` " + set + " \"" + val + "\"")
				}	
			} else {
				clauses = append(clauses, "`" + col + "`" + comparator + "\"" + val + "\"")
			}
		}
	}
	return clauses
} 

func collectSet(r *http.Request, cred Access, db string, t string) string {
	return strings.Join(collectClauses (r, cred, db, t, "="), " , ")
}

func collectWhere(r *http.Request, cred Access, db string, t string) string {
	return strings.Join(collectClauses (r, cred, db, t, "")," && ")
}
 

func actionQuery(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {

	where := collectWhere(r, cred, db, t)
	if len(where) > 0 {
		query := "SELECT * FROM `" + t + "` WHERE " + where
		dumpRows(w, db, t, "", "", cred, query, where)
	}
}

func actionDeleteExec(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {

	where := collectWhere(r, cred, db, t)
	if len(where) > 0 {
		stmt := "DELETE FROM `" + t + "` WHERE " + where
		log.Println("[SQL]", stmt)
		conn := getConnection(cred, db)
		defer conn.Close()

		statement, err := conn.Prepare(stmt)
		checkY(err)
		_, err = statement.Exec()
		if err != nil {
			shipError(w, cred, db, t, stmt, err)
		}
		http.Redirect(w, r, r.URL.Host+"?db="+db+"&t="+t, 302)
	}
}

func actionDeleteWhere(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, where string) {

	stmt := "DELETE FROM `" + t + "` WHERE " + where
	log.Println("[SQL]", stmt)
	conn := getConnection(cred, db)
	defer conn.Close()

	statement, err := conn.Prepare(stmt)
	checkY(err)
	_, err = statement.Exec()
	if err != nil {
		shipError(w, cred, db, t, stmt, err)
	}
	http.Redirect(w, r, r.URL.Host+"?db="+db+"&t="+t, 302)
}

func actionInsert(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {

	clauses := collectSet(r, cred, db, t)
	if len(clauses) > 0 {
		stmt := "INSERT INTO `" + t + "` SET " + clauses
		log.Println("[SQL]", stmt)
		conn := getConnection(cred, db)
		defer conn.Close()

		statement, err := conn.Prepare(stmt)
		checkY(err)
		_, err = statement.Exec()
		if err != nil {
			shipError(w, cred, db, t, stmt, err)
		}
		http.Redirect(w, r, r.URL.Host+"?db="+db+"&t="+t, 302)
	}
}

func actionEditExec(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, k string, v string) {

	clauses := collectSet(r, cred, db, t)
	if len(clauses) > 0 {
		stmt := "UPDATE `" + t + "` SET " + clauses + " WHERE `" + k + "` = ?"
		log.Println("[SQL]", stmt, v)
		conn := getConnection(cred, db)
		defer conn.Close()

		statement, err := conn.Prepare(stmt)
		checkY(err)
		_, err = statement.Exec(v)
		if err != nil {
			shipError(w, cred, db, t, stmt, err)
		}
		http.Redirect(w, r, r.URL.Host+"?db="+db+"&t="+t, 302)
	}
}

func actionRemove(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, k string, v string) {

	stmt := "DELETE FROM `" + t + "` WHERE `" + k + "` = ?"
	log.Println("[SQL]", stmt, v)
	conn := getConnection(cred, db)
	defer conn.Close()

	statement, err := conn.Prepare(stmt)
	checkY(err)
	_, err = statement.Exec(v)
	if err != nil {
		shipError(w, cred, db, t, stmt, err)
	}
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

	rows, err := getRows(cred, db, "show columns from `" + t + "`")
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
