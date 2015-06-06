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
	"html"
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
	Back     string
	Columns  []CContext
	Hidden   []CContext
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
		Hidden:	  []CContext{},
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

func shipForm(w http.ResponseWriter, r *http.Request, cred Access, 
              db string, t string,
              action string, button string, selector string, 
              vmap map[string]string, hiddencols []CContext) {

	cols := getColumnInfo(cred, db, t)
	primary := getPrimary(cred, db, t)
	newcols := []CContext{}

	for _, col := range cols {
		name := html.EscapeString(col.Name)
		readonly := ""
		value := html.EscapeString(vmap[col.Name])
		label := ""
		if name == primary {
			label = name + " (ID)"
			readonly = value
		} else {
			label = name
		}
		newcols = append(newcols, CContext{col.Number, name, label, col.IsNumeric, col.IsString, value, readonly})
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
		Back:     linkback,
		Columns:  newcols,
		Hidden:	  hiddencols,
		Trail:    makeTrail(cred.Host, db, t, "", "", "", ""),
	}

	if DEBUGFLAG {
		initTemplate()
	}
	err := templateFormFields.Execute(w, c)
	checkY(err)
}

func actionSubset(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {
	shipForm(w, r, cred, db, t, "QUERY", "Query", "true", make(map[string]string),[]CContext{})
}

func actionDeleteForm(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {
	shipForm(w, r, cred, db, t, "DELETEEXEC", "Delete", "true", make(map[string]string),[]CContext{})
}

func actionAdd(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {
	shipForm(w, r, cred, db, t, "INSERT", "Insert", "", make(map[string]string),[]CContext{})
}

func actionUpdateSubset(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {
	cols := getCols(cred, db, t)
	wclauses,_ ,whereQ:= collectClauses(r, cols)
	where := strings.Join(wclauses, " && ")
	hiddencols:=[]CContext{}
	for field, valueArray := range whereQ {    //type Values map[string][]string
		hiddencols = append(hiddencols, CContext{"", field, "", "", "", valueArray[0], ""})
	}
	
	count := getSingleValue(cred, db, "select count(*) from `" + t + "` where " + where)
	if count== "1" {
		rows,err := getRows(cred, db, "select * from `" + t + "` where " + where)
		checkY(err)
		shipForm(w, r, cred, db, t, "UPDATEEXEC", "Update", "", getValueMap(w, db, t, cred, rows), hiddencols)
	} else {
		shipForm(w, r, cred, db, t, "UPDATEEXEC", "Update", "", make(map[string]string), hiddencols)
	}
}

func actionEdit(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, k string, v string) {
	hiddencols:=[]CContext{CContext{"", "k", "", "", "",k, ""},CContext{"", "v", "", "", "", v, ""}}
	query := "select * from `" + t + "` where `" + k + "` = ?"
	conn := getConnection(cred, db)
	defer conn.Close()
	stmt, err := conn.Prepare(query)
	checkY(err)
	rows, err := stmt.Query(v)
	checkY(err)
	shipForm(w, r, cred, db, t, "EDITEXEC", "Submit", "", getValueMap(w, db, t, cred, rows),hiddencols)
}

func collectClauses(r *http.Request, cols []string) ([]string, []string, url.Values) {

	v := url.Values{}
	var whereclauses,setclauses []string
	for _, col := range cols {
		colname := sqlProtectIdentifier(col)
		colhtml := html.EscapeString(col)
		valraw := r.FormValue(col + "W")
		setraw := r.FormValue(col + "S")
		val := sqlProtectString(valraw)
		set := sqlProtectString(setraw)
		if val != "" {
			v.Add(colhtml+"W",valraw)
			comparaw := r.FormValue(col + "O")
			comparator := sqlProtectString(comparaw)
			if comparator == "" {
				whereclauses = append(whereclauses, "`"+colname+"`"+sqlProtectNumericComparison(val))
			} else if comparator=="~"{
				v.Add(colhtml+"O",comparaw)
				whereclauses = append(whereclauses, "`"+colname+"` LIKE \""+val+"\"")
			} else if comparator=="!~"{
				v.Add(colhtml+"O",comparaw)
				whereclauses = append(whereclauses, "`"+colname+"` NOT LIKE \""+val+"\"")
			} else {
				v.Add(colhtml+"O",sqlProtectNumericComparison(comparaw))
				whereclauses = append(whereclauses, "`"+colname+"` "+sqlProtectNumericComparison(comparaw)+" \""+val+"\"")
			}
		} 
		if set != "" {
			v.Add(colhtml+"S",setraw)
			setclauses = append(setclauses, "`"+colname+"` "+"="+" \""+set+"\"")
		}
	}
	return whereclauses,setclauses,v
}

// TODO: submit reader to collectClauses and unify	
func WhereQuery2Sql(q url.Values, cols []string) string {
	
	var clauses []string
	log.Println("WhereQ2S",q)
	for _, col := range cols {
		colname := sqlProtectIdentifier(col)
		val := sqlProtectString(q.Get(html.EscapeString(col) + "W"))
		if val != "" {
			comparator := sqlProtectString(q.Get(html.EscapeString(col) + "O"))
			if comparator == "" {
				clauses = append(clauses, "`"+colname+"`"+sqlProtectNumericComparison(val))
			} else if comparator=="~"{
				clauses = append(clauses, "`"+colname+"` LIKE \""+val+"\"")
			} else if comparator=="!~"{
				clauses = append(clauses, "`"+colname+"` NOT LIKE \""+val+"\"")
			} else {
				clauses = append(clauses, "`"+colname+"` "+sqlProtectNumericComparison(comparator)+" \""+val+"\"")
			}
		}
	}
	log.Println("WhereQ2S",strings.Join(clauses," && "))
	return strings.Join(clauses," && ")
}	

func actionQuery(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {

	cols := getCols(cred, db, t)
	wclauses,_ ,whereQ:= collectClauses(r, cols)
	where := strings.Join(wclauses, " && ")
	if len(where) > 0 {
		query := "SELECT * FROM `" + t + "` WHERE " + where
		dumpRows(w, db, t, "", "", cred, query, whereQ)
	}
}

func actionUpdateExec(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {

	cols := getCols(cred, db, t)
	wclauses,sclauses,_:= collectClauses(r, cols)
	sets  := strings.Join(sclauses, " , ")
	where := strings.Join(wclauses, " && ")
	if len(sclauses) > 0 {
		stmt := "UPDATE `" + t + "` SET " + sets + " WHERE " + where
		log.Println("[SQL]", stmt)
		conn := getConnection(cred, db)
		defer conn.Close()

		statement, err := conn.Prepare(stmt)
		if err != nil {
			shipError(w, cred, db, t, stmt, err)
		}
		_, err = statement.Exec()
		if err != nil {
			shipError(w, cred, db, t, stmt, err)
		}
		http.Redirect(w, r, r.URL.Host+"?db="+db+"&t="+t, 302)
	}
}


func actionDeleteExec(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {

	cols := getCols(cred, db, t)
	wclauses,_,_:= collectClauses(r, cols)
	where := strings.Join(wclauses, " && ")
	if len(where) > 0 {
		stmt := "DELETE FROM `" + t + "` WHERE " + where
		log.Println("[SQL]", stmt)
		conn := getConnection(cred, db)
		defer conn.Close()

		statement, err := conn.Prepare(stmt)
		if err != nil {
			shipError(w, cred, db, t, stmt, err)
		}
		_, err = statement.Exec()
		if err != nil {
			shipError(w, cred, db, t, stmt, err)
		}
		http.Redirect(w, r, r.URL.Host+"?db="+db+"&t="+t, 302)
	}
}

func actionDeleteSubset(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {
	
	cols := getCols(cred, db, t)
	wclauses,_,_:= collectClauses(r, cols)
	where := strings.Join(wclauses, " && ")
	if len(where) > 0 {
		stmt := "DELETE FROM `" + t + "` WHERE " + where
		log.Println("[SQL]", stmt)
		conn := getConnection(cred, db)
		defer conn.Close()

		statement, err := conn.Prepare(stmt)
		if err != nil {
			shipError(w, cred, db, t, stmt, err)
		}
		_, err = statement.Exec()
		if err != nil {
			shipError(w, cred, db, t, stmt, err)
		}
		http.Redirect(w, r, r.URL.Host+"?db="+db+"&t="+t, 302)
	}
}

func collectSet(r *http.Request, cred Access, db string, t string) string {
	cols := getCols(cred, db, t)
	_,sclauses,_ := collectClauses(r, cols)
	return strings.Join(sclauses, " , ")
}

func actionInsert(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {

	clauses := collectSet(r, cred, db, t)
	if len(clauses) > 0 {
		stmt := "INSERT INTO `" + t + "` SET " + clauses
		log.Println("[SQL]", stmt)
		conn := getConnection(cred, db)
		defer conn.Close()

		statement, err := conn.Prepare(stmt)
		if err != nil {
			shipError(w, cred, db, t, stmt, err)
		}
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
		if err != nil {
			shipError(w, cred, db, t, stmt, err)
		}
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
	if err != nil {
		shipError(w, cred, db, t, stmt, err)
	}
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

	rows, err := getRows(cred, db, "show columns from `"+t+"`")
	checkY(err)
	defer rows.Close()

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)

	menu := []Entry{}
	q.Set("action", "ADD")
	linkinsert := q.Encode()
	menu = append(menu, Entry{"+", linkinsert})

	records := [][]Entry{}
	head := []Entry{{"#", ""}, {"Field", ""}, {"Type", ""}, {"Null", ""}, {"Key", ""}, {"Default", ""}, {"Extra", ""}}

	var i int = 1
	for rows.Next() {
		var f, t, n, k, e string
		var d []byte // or use http://golang.org/pkg/database/sql/#NullString
		err := rows.Scan(&f, &t, &n, &k, &d, &e)
		checkY(err)
		records = append(records, []Entry{{strconv.Itoa(i), ""}, escape(f, ""), {t, ""}, {n, ""}, {k, ""}, {string(d), ""}, {e, ""}})
		i = i + 1
	}
	tableOutSimple(w, cred, db, t, head, records, menu)
}
