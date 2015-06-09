package main

/*
<form  action="/login">
   <label for="user">User</label><input type="text"     id="user" name="user"><br>
   <label for="pass">Pass</label><input type="password" id="pass" name="pass"><br>
   <label for="host">Host</label><input type="text"     id="host" name="host" value="localhost"><br>
   <label for="port">Port</label><input type="text"     id="port" name="port" value="3306"><br>
   <button type="submit">Select</button>
</form>
*/

import (
	_ "github.com/go-sql-driver/mysql"
	"html"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type FContext struct {
	CSS      string
	Action   string
	Selector string
	Button   string
	Database string
	Table    string
	Order    string
	Desc     string
	Back     string
	Columns  []CContext
	Hidden   []CContext
	Trail    []Entry
}

func shipForm(w http.ResponseWriter, r *http.Request, cred Access,
	db string, t string, o string, d string,
	action string, button string, selector string, vmap map[string]string, hiddencols []CContext) {

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
		Order:    o,
		Desc:     d,
		Back:     linkback,
		Columns:  newcols,
		Hidden:   hiddencols,
		Trail:    makeTrail(cred.Host, db, t, o, d, "", "", url.Values{}),
	}

	if DEBUGFLAG {
		initTemplate()
	}
	err := templateFormFields.Execute(w, c)
	checkY(err)
}

func actionQUERY(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, o string, d string) {
	shipForm(w, r, cred, db, t, o, d, "SELECT", "Select", "true", make(map[string]string), []CContext{})
}

func actionDELETEFORM(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, o string, d string) {
	shipForm(w, r, cred, db, t, o, d, "DELETEEXEC", "Delete", "true", make(map[string]string), []CContext{})
}

func actionADD(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, o string, d string) {
	shipForm(w, r, cred, db, t, o, d, "INSERT", "Insert", "", make(map[string]string), []CContext{})
}

func actionUPDATE(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, o string, d string) {
	cols := getCols(cred, db, t)
	wclauses, _, whereQ := collectClauses(r, cols)
	where := strings.Join(wclauses, " && ")
	hiddencols := []CContext{}
	for field, valueArray := range whereQ { //type Values map[string][]string
		hiddencols = append(hiddencols, CContext{"", field, "", "", "", valueArray[0], ""})
	}

	count := getSingleValue(cred, db, "select count(*) from `"+t+"` where "+where)
	if count == "1" {
		rows, err := getRows(cred, db, "select * from `"+t+"` where "+where)
		checkY(err)
		shipForm(w, r, cred, db, t, o, d, "UPDATEEXEC", "Update", "", getValueMap(w, db, t, cred, rows), hiddencols)
	} else {
		shipForm(w, r, cred, db, t, o, d, "UPDATEEXEC", "Update", "", make(map[string]string), hiddencols)
	}
}


// TODO: to allow for submitting multiple clauses for a field, they should be numbered W1, O1 ...
func collectClauses(r *http.Request, cols []string) ([]string, []string, url.Values) {

	v := url.Values{}
	var whereclauses, setclauses []string
	for _, col := range cols {
		colname := sqlProtectIdentifier(col)
		colhtml := html.EscapeString(col)
		val := r.FormValue(col + "W")
		set := r.FormValue(col + "S")
		if val != "" {
			v.Add(colhtml+"W", val)
			comp := r.FormValue(col + "O")
			if comp == "" {
				comp, val = sqlFilterNumericComparison(val)
				whereclauses = append(whereclauses, "`"+colname+"`"+sqlFilterComparator(comp)+"'"+sqlFilterNumber(val)+"'")
			} else if comp == "~" {
				v.Add(colhtml+"O", comp)
				whereclauses = append(whereclauses, "`"+colname+"` LIKE \""+sqlProtectString(val)+"\"")
			} else if comp == "!~" {
				v.Add(colhtml+"O", comp)
				whereclauses = append(whereclauses, "`"+colname+"` NOT LIKE \""+sqlProtectString(val)+"\"")
			} else if comp == "==" {
				v.Add(colhtml+"O", comp)
				whereclauses = append(whereclauses, "BINARY `"+colname+"`=\""+sqlProtectString(val)+"\"")
			} else if comp == "!=" {
				v.Add(colhtml+"O", comp)
				whereclauses = append(whereclauses, "BINARY `"+colname+"`!=\""+sqlProtectString(val)+"\"")
			} else {
				v.Add(colhtml+"O", comp)
				if sqlFilterNumber(val) != "" {
					whereclauses = append(whereclauses, "`"+colname+"`"+sqlFilterComparator(comp)+"'"+sqlFilterNumber(val)+"'")
				} else {
					whereclauses = append(whereclauses, "`"+colname+"`"+sqlFilterComparator(comp)+"\""+sqlProtectString(val)+"\"")
				}
			}
		}
		if set != "" {
			v.Add(colhtml+"S", set)
			setclauses = append(setclauses, "`"+colname+"` "+"="+" \""+sqlProtectString(set)+"\"")
		}
	}
	return whereclauses, setclauses, v
}

func WhereSelect2Pretty(q url.Values, ccols []CContext) string {

	var clauses []string
	for _, col := range ccols {
		colname := col.Label
		val := q.Get(html.EscapeString(col.Name) + "W")
		if val != "" {
			comp := q.Get(html.EscapeString(col.Name) + "O")
			if comp == "" {
				comp, val = sqlFilterNumericComparison(val)
				clauses = append(clauses, colname+sqlFilterComparator(comp)+sqlFilterNumber(val))
			} else if comp == "~" {
				clauses = append(clauses, colname + " LIKE \"" + val + "\"")
			} else if comp == "!~" {
				clauses = append(clauses, colname + " NOT LIKE \"" + val + "\"")
			} else if comp == "==" {
				clauses = append(clauses, colname + "==\"" + val + "\"")
			} else if comp == "!=" {
				clauses = append(clauses, colname + "!=\"" + val + "\"")
			} else {
				if col.IsNumeric != "" {
					clauses = append(clauses, colname+sqlFilterComparator(comp)+sqlFilterNumber(val))
				} else {
					clauses = append(clauses, colname+sqlFilterComparator(comp)+" \""+val+"\"")
				}
			}
		}
	}
	return html.EscapeString(strings.Join(clauses, " AND "))
}

func actionSELECT(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, o string, d string) {

	cols := getCols(cred, db, t)
	wclauses, _, whereQ := collectClauses(r, cols)
	where := strings.Join(wclauses, " && ")
	if where != "" {
		query := "Select * FROM `" + t + "` WHERE " + where
		dumpRows(w, db, t, o, d, cred, query, whereQ)
	}
}

func actionUPDATEEXEC(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, o string, d string) {

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("o", o)
	q.Add("d", o)
	cols := getCols(cred, db, t)
	wclauses, sclauses, _ := collectClauses(r, cols)
	sets := strings.Join(sclauses, " , ")
	where := strings.Join(wclauses, " && ")
	if len(sclauses) > 0 {
		stmt := "UPDATE `" + t + "` SET " + sets + " WHERE " + where
		log.Println("[SQL]", stmt)
		conn := getConnection(cred, db)
		defer conn.Close()

		preparedStmt, err := conn.Prepare(stmt)
		checkErrorPage(w, cred, db, t, stmt, err)
		_, err = preparedStmt.Exec()
		checkErrorPage(w, cred, db, t, stmt, err)
		http.Redirect(w, r, "?"+q.Encode(), 302)
	}
}

func actionDELETEEXEC(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, o string, d string) {

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("o", o)
	q.Add("d", d)
	cols := getCols(cred, db, t)
	wclauses, _, _ := collectClauses(r, cols)
	where := strings.Join(wclauses, " && ")
	if len(where) > 0 {
		stmt := "DELETE FROM `" + t + "` WHERE " + where
		log.Println("[SQL]", stmt)
		conn := getConnection(cred, db)
		defer conn.Close()

		preparedStmt, err := conn.Prepare(stmt)
		checkErrorPage(w, cred, db, t, stmt, err)
		_, err = preparedStmt.Exec()
		checkErrorPage(w, cred, db, t, stmt, err)
		http.Redirect(w, r, "?"+q.Encode(), 302)
	}
}

func actionDELETE(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, o string, d string) {

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("o", o)
	q.Add("d", d)
	cols := getCols(cred, db, t)
	wclauses, _, _ := collectClauses(r, cols)
	where := strings.Join(wclauses, " && ")
	if len(where) > 0 {
		stmt := "DELETE FROM `" + t + "` WHERE " + where
		conn := getConnection(cred, db)
		defer conn.Close()

		log.Println("[SQL]", stmt)
		preparedStmt, err := conn.Prepare(stmt)
		checkErrorPage(w, cred, db, t, stmt, err)
		_, err = preparedStmt.Exec()
		checkErrorPage(w, cred, db, t, stmt, err)
		http.Redirect(w, r, "?"+q.Encode(), 302)
	}
}

func collectSet(r *http.Request, cred Access, db string, t string) string {
	cols := getCols(cred, db, t)
	_, sclauses, _ := collectClauses(r, cols)
	return strings.Join(sclauses, " , ")
}

func actionINSERT(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, o string, d string) {

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("o", o)
	q.Add("d", d)
	clauses := collectSet(r, cred, db, t)
	if len(clauses) > 0 {
		stmt := "INSERT INTO `" + t + "` SET " + clauses
		log.Println("[SQL]", stmt)
		conn := getConnection(cred, db)
		defer conn.Close()

		preparedStmt, err := conn.Prepare(stmt)
		checkErrorPage(w, cred, db, t, stmt, err)
		_, err = preparedStmt.Exec()
		checkErrorPage(w, cred, db, t, stmt, err)
		http.Redirect(w, r, "?"+q.Encode(), 302)
	}
}

/* The next three functions deal with modifications in tables with primary key:
 * They use prepared statements.
 * However, these tempates only deal with values, not with identifiers. */


func actionEDIT(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, k string, v string) {
	hiddencols := []CContext{CContext{"", "k", "", "", "", k, ""}, CContext{"", "v", "", "", "", v, ""}}
	stmt := "select * from `" + t + "` where `" + k + "`=?"
	conn := getConnection(cred, db)
	defer conn.Close()

	log.Println("[SQL]", stmt, " <= ", v)
	preparedStmt, err := conn.Prepare(stmt)
	checkY(err)
	rows, err := preparedStmt.Query(v)
	checkY(err)
	shipForm(w, r, cred, db, t, "", "", "EDITEXEC", "Submit", "", getValueMap(w, db, t, cred, rows), hiddencols)
}

func actionEDITEXEC(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, k string, v string) {
	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	clauses := collectSet(r, cred, db, t)
	if len(clauses) > 0 {
		stmt := "UPDATE `" + t + "` SET " + clauses + " WHERE `" + k + "` = ?"
		conn := getConnection(cred, db)
		defer conn.Close()

		log.Println("[SQL]", stmt, " <= ", v)
		preparedStmt, err := conn.Prepare(stmt)
		checkErrorPage(w, cred, db, t, stmt, err)
		_, err = preparedStmt.Exec(v)
		checkErrorPage(w, cred, db, t, stmt, err)
		http.Redirect(w, r, "?"+q.Encode(), 302)
	}
}

func actionREMOVE(w http.ResponseWriter, r *http.Request, cred Access, db string, t string, k string, v string) {
	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	stmt := "DELETE FROM `" + t + "` WHERE `" + k + "` = ?"
	conn := getConnection(cred, db)
	defer conn.Close()

	log.Println("[SQL]", stmt, " <= ", v)
	preparedStmt, err := conn.Prepare(stmt)
	checkErrorPage(w, cred, db, t, stmt, err)
	_, err = preparedStmt.Exec(v)
	checkErrorPage(w, cred, db, t, stmt, err)
	http.Redirect(w, r, "?"+q.Encode(), 302)
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

func actionINFO(w http.ResponseWriter, r *http.Request, cred Access, db string, t string) {

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
