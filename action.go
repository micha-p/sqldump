package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"html"
	"net/http"
	"net/url"
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

// INSERTFORM and SELECTFORM provide columns without values, EDIT/UPDATE provide a filled vmap
// TODO: use DEFAULT and AUTOINCREMENT

func shipForm(w http.ResponseWriter, r *http.Request, conn *sql.DB,
	host string, db string, t string, o string, d string,
	action string, button string, selector string, showncols []CContext, hiddencols []CContext) {

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
		Columns:  showncols,
		Hidden:   hiddencols,
		Trail:    makeTrail(host, db, t, "", url.Values{}),
	}

	if DEBUGFLAG {
		initTemplate()
	}
	err := templateForm.Execute(w, c)
	checkY(err)
}

/* The next four functions generate forms for doing SELECT, DELETE, INSERT, UPDATE */

func actionSELECTFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, o string, d string) {
	shipForm(w, r, conn, host, db, t, o, d, "SELECT", "Select", "true", getColumnInfo(conn, t), []CContext{})
}

func actionDELETEFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, o string, d string) {
	shipForm(w, r, conn, host, db, t, o, d, "DELETE", "Delete", "true", getColumnInfo(conn, t), []CContext{})
}

func actionINSERTFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, o string, d string) {
	shipForm(w, r, conn, host, db, t, o, d, "INSERT", "Insert", "", getColumnInfo(conn, t), []CContext{})
}

func actionUPDATEFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, o string, d string) {

	wclauses, _, whereQ := collectClauses(r, conn, t)
	hiddencols := []CContext{}
	for field, valueArray := range whereQ { //type Values map[string][]string
		hiddencols = append(hiddencols, CContext{"", field, "", "", "", "", "valid", valueArray[0], ""})
	}

	count, _ := getSingleValue(conn, host, db, sqlCount(t)+sqlWhereClauses(wclauses))
	if count == "1" {
		rows, err, _ := getRows(conn, sqlStar(t)+sqlWhereClauses(wclauses))
		checkY(err)
		defer rows.Close()
		shipForm(w, r, conn, host, db, t, o, d, "UPDATE", "Update", "", getColumnInfoFilled(conn, host, db, t, "", rows), hiddencols)
	} else {
		shipForm(w, r, conn, host, db, t, o, d, "UPDATE", "Update", "", getColumnInfo(conn, t), hiddencols)
	}
}

func actionKV_UPDATEFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, k string, v string) {
	hiddencols := []CContext{
		CContext{"", "k", "", "", "", "", "valid", k, ""},
		CContext{"", "v", "", "", "", "", "valid", v, ""}}
	stmt := sqlStar(t) + sqlWhere1(k, "=")
	preparedStmt, _, err := sqlPrepare(conn, stmt)
	defer preparedStmt.Close()
	checkErrorPage(w, host, db, t, stmt, err)
	rows, _,err := sqlQuery1(preparedStmt,v)
	checkY(err)
	defer rows.Close()
	primary := getPrimary(conn, t)
	shipForm(w, r, conn, host, db, t, "", "", "KV_UPDATE", "Update", "", getColumnInfoFilled(conn, host, db, t, primary, rows), hiddencols)
}

func actionGV_UPDATEFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, g string, v string) {
	hiddencols := []CContext{
		CContext{"", "g", "", "", "", "", "valid", g, ""},
		CContext{"", "v", "", "", "", "", "valid", v, ""}}
	stmt := sqlStar(t) + sqlWhere1(g, "=")
	preparedStmt, _, err := sqlPrepare(conn, stmt)
	defer preparedStmt.Close()
	checkErrorPage(w, host, db, t, stmt, err)
	rows, _,err := sqlQuery1(preparedStmt,v)
	checkY(err)
	defer rows.Close()
	primary := getPrimary(conn, t)
	shipForm(w, r, conn, host, db, t, "", "", "GV_UPDATE", "Update", "", getColumnInfoFilled(conn, host, db, t, primary, rows), hiddencols)
}


// TODO: to allow for submitting multiple clauses for a field, they should be numbered W1, O1 ...
func collectClauses(r *http.Request, conn *sql.DB, t string) ([]sqlstring, []sqlstring, url.Values) {

	v := url.Values{}
	// we have to retrieve columns for getting where clauses => TODO: w1=column,c1,v1=value,s1=column
	cols := getCols(conn, t)
	var whereclauses, setclauses []sqlstring
	for _, col := range cols {
		colname := sqlProtectIdentifier(col)
		colhtml := html.EscapeString(col)
		val := r.FormValue(col + "W")
		set := r.FormValue(col + "S")
		null := r.FormValue(col + "N")
		comp := r.FormValue(col + "O")
		if val != "" || comp == "=0" || comp == "!0" {
			v.Add(colhtml+"W", val)
			if comp == "" {
				comp, val = sqlFilterNumericComparison(val)
				whereclauses = append(whereclauses, colname+sqlFilterComparator(comp)+sqlFilterNumber(val))
			} else if comp == "=" {
				v.Add(colhtml+"O", comp)
				whereclauses = append(whereclauses, colname+string2sql(" = ")+sqlProtectString(val))
			} else if comp == "~" {
				v.Add(colhtml+"O", comp)
				whereclauses = append(whereclauses, colname+string2sql(" LIKE ")+sqlProtectString(val))
			} else if comp == "!~" {
				v.Add(colhtml+"O", comp)
				whereclauses = append(whereclauses, colname+string2sql(" NOT LIKE ")+sqlProtectString(val))
			} else if comp == "==" {
				v.Add(colhtml+"O", comp)
				whereclauses = append(whereclauses, string2sql("BINARY ")+colname+string2sql("=")+sqlProtectString(val))
			} else if comp == "!=" {
				v.Add(colhtml+"O", comp)
				whereclauses = append(whereclauses, string2sql("BINARY ")+colname+string2sql("!=")+sqlProtectString(val))
			} else if comp == "=0" {
				v.Add(colhtml+"O", comp)
				whereclauses = append(whereclauses, colname+string2sql(" IS NULL"))
			} else if comp == "!0" {
				v.Add(colhtml+"O", comp)
				whereclauses = append(whereclauses, colname+string2sql(" IS NOT NULL"))
			} else {
				v.Add(colhtml+"O", comp)
				if sqlFilterNumber(val) != "" {
					whereclauses = append(whereclauses, colname+sqlFilterComparator(comp)+sqlFilterNumber(val))
				} else {
					whereclauses = append(whereclauses, colname+sqlFilterComparator(comp)+sqlProtectString(val))
				}
			}
		}
		if null != "" {
			v.Add(colhtml+"N", null)
			setclauses = append(setclauses, colname+"=NULL")
		} else if set != "" {
			v.Add(colhtml+"S", set)
			setclauses = append(setclauses, colname+"="+sqlProtectString(set))
		} else {
			setclauses = append(setclauses, colname+"="+"\"\"")
		}
	}
	return whereclauses, setclauses, v
}

func WhereQuery2Pretty(q url.Values, ccols []CContext) string {
	var clauses []string
	for _, col := range ccols {
		colname := col.Label
		val := q.Get(html.EscapeString(col.Name) + "W")
		comp := q.Get(html.EscapeString(col.Name) + "O")
		if val != "" || comp == "=0" || comp == "!0" {
			if comp == "" {
				comp, val = sqlFilterNumericComparison(val)
				clauses = append(clauses, colname+sql2string(sqlFilterComparator(comp))+sql2string(sqlFilterNumber(val)))
			} else if comp == "~" {
				clauses = append(clauses, colname+" LIKE \""+val+"\"")
			} else if comp == "!~" {
				clauses = append(clauses, colname+" NOT LIKE \""+val+"\"")
			} else if comp == "==" {
				clauses = append(clauses, colname+"==\""+val+"\"")
			} else if comp == "!=" {
				clauses = append(clauses, colname+"!=\""+val+"\"")
			} else if comp == "=0" {
				clauses = append(clauses, colname+" IS NULL")
			} else if comp == "!0" {
				clauses = append(clauses, colname+" IS NOT NULL")
			} else {
				if col.IsNumeric != "" {
					clauses = append(clauses, colname+sql2string(sqlFilterComparator(comp))+sql2string(sqlFilterNumber(val)))
				} else {
					clauses = append(clauses, colname+sql2string(sqlFilterComparator(comp))+" \""+val+"\"")
				}
			}
		}
	}
	return strings.Join(clauses, " & ")
}


// Excutes a statement on a selection by where-clauses
// Used, when rows are not adressable by a primary key or in table having a group
func actionEXEC(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, stmt sqlstring) {

	messageStack := []Message{}
	preparedStmt, sec, err := sqlPrepare(conn, stmt)
	defer preparedStmt.Close()
	checkErrorPage(w, host, db, t, stmt, err)
	messageStack = append(messageStack, Message{"PREPARE stmt FROM '" + sql2string(stmt) + "'", -1,0,sec})

	result, sec, err := sqlExec(preparedStmt)
	checkErrorPage(w, host, db, t, stmt, err)
	affected, err := result.RowsAffected()
	checkErrorPage(w, host, db, t, stmt, err)

	messageStack = append(messageStack, Message{"EXECUTE stmt", -1,affected,sec})
	nextstmt := sqlStar(t) + sqlOrder(o,d)
	dumpRows(w, conn, host, db, t, o, d, messageStack, nextstmt)
}


/* Executes prepared statements about modifications in tables with primary key or having a group
 * Uses one argument as value for where clause
 * However, prepared statements only work with values, not in identifier position */
func actionEXEC1(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, stmt sqlstring, arg string) {

	messageStack := []Message{}
	preparedStmt, sec, err := sqlPrepare(conn, stmt)
	defer preparedStmt.Close()
	checkErrorPage(w, host, db, t, stmt, err)
	messageStack = append(messageStack, Message{"PREPARE stmt FROM '" + sql2string(stmt) + "'", -1,0,sec})

	result, sec, err := sqlExec1(preparedStmt,arg)
	checkErrorPage(w, host, db, t, stmt, err)
	affected, err := result.RowsAffected()
	checkErrorPage(w, host, db, t, stmt, err)

	messageStack = append(messageStack, Message{"EXECUTE stmt USING \"" + arg + "\"", -1,affected,sec})
	nextstmt := sqlStar(t)
	dumpRows(w, conn, host, db, t, "", "", messageStack, nextstmt)
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
func actionINFO(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string) {

	stmt := sqlColumns(t)
	rows, err, _:= getRows(conn, stmt)
	checkY(err)
	defer rows.Close()

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Set("action", "SELECTFORM")
	linkselect := q.Encode()
	q.Set("action", "INSERTFORM")
	linkinsert := q.Encode()
	q.Set("action", "DELETEFORM")
	linkdeleteF := q.Encode()
	q.Set("action", "INFO")
	linkinfo := q.Encode()
	q.Del("action")

	menu := []Entry{}
	menu = append(menu, escape("?", linkselect))
	menu = append(menu, escape("+", linkinsert))
	menu = append(menu, escape("-", linkdeleteF))
	menu = append(menu, escape("i", linkinfo))

	records := [][]Entry{}
	head := []Entry{escape("#"), escape("Field"), escape("Type"), escape("Null"), escape("Key"), escape("Default"), escape("Extra")}

	var i int64 = 1
	for rows.Next() {
		var f, t, n, k, e string
		var d []byte // or use http://golang.org/pkg/database/sql/#NullString
		err := rows.Scan(&f, &t, &n, &k, &d, &e)
		checkY(err)
		records = append(records, []Entry{escape(Int64toa(i)), escape(f), escape(t), escape(n), escape(k), escape(string(d)), escape(e)})
		i = i + 1
	}
	// message not shown as it disturbs equal alignment of info, query and field.
	tableOutSimple(w, conn, host, db, t, head, records, menu)
}
