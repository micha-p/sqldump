package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"html"
	"net/http"
	"net/url"
	"strings"
)

// TODO: show results page like rows page with rows_affected

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

func actionSELECT(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, o string, d string) {
	wclauses, _, whereQ := collectClauses(r, conn, t)
	if len(wclauses) > 0 {
		query := sqlStar(t) + sqlWhereClauses(wclauses)
		dumpWhere(w, conn, host, db, t, o, d, query, whereQ)
	} else {
		shipMessage(w, host, db, "Where clauses not found")
	}
}

// the next three functions deal with tables where a primary key is not existant or not in use

func actionINSERT(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, o string, d string) {

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("o", o)
	q.Add("d", d)
	_, sclauses, _ := collectClauses(r, conn, t)
	if len(sclauses) > 0 {
		stmt := sqlInsert(t) + sqlSetClauses(sclauses)
		preparedStmt, err := sqlPrepare(conn, stmt)
		defer preparedStmt.Close()
		checkErrorPage(w, host, db, t, stmt, err)
		result, sec, err := sqlExec(preparedStmt)
		checkErrorPage(w, host, db, t, stmt, err)
		affected, err := result.RowsAffected()
		checkErrorPage(w, host, db, t, stmt, err)
		nextstmt := sqlStar(t)
		dumpQuery(w, conn, host, db, t, o, d, []Message{Message{Msg:sql2string(stmt), Rows: -1, Affected:affected, Seconds: sec}}, nextstmt)
	} else {
		shipMessage(w, host, db, "Set clauses not found")
	}
}

func actionUPDATE(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, o string, d string) {

	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("o", o)
	q.Add("d", o)
	wclauses, sclauses, _ := collectClauses(r, conn, t)
	if len(sclauses) > 0 {
		stmt := sqlUpdate(t) + sqlSetClauses(sclauses) + sqlWhereClauses(wclauses)
		preparedStmt, err := sqlPrepare(conn, stmt)
		defer preparedStmt.Close()
		checkErrorPage(w, host, db, t, stmt, err)
		result, sec, err := sqlExec(preparedStmt)
		checkErrorPage(w, host, db, t, stmt, err)
		affected, err := result.RowsAffected()
		checkErrorPage(w, host, db, t, stmt, err)
		nextstmt := sqlStar(t)
		dumpQuery(w, conn, host, db, t, o, d, []Message{Message{Msg:sql2string(stmt), Rows: -1, Affected:affected, Seconds: sec}}, nextstmt)
	} else {
		shipMessage(w, host, db, "Set clauses not found")
	}
}

func actionDELETE(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, o string, d string) {
	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	q.Add("o", o)
	q.Add("d", d)
	wclauses, _, _ := collectClauses(r, conn, t)
	if len(wclauses) > 0 {
		stmt := sqlDelete(t) + sqlWhereClauses(wclauses)
		preparedStmt, err := sqlPrepare(conn, stmt)
		defer preparedStmt.Close()
		checkErrorPage(w, host, db, t, stmt, err)
		result, sec, err := sqlExec(preparedStmt)
		checkErrorPage(w, host, db, t, stmt, err)
		affected, err := result.RowsAffected()
		checkErrorPage(w, host, db, t, stmt, err)
		nextstmt := sqlStar(t)
		dumpQuery(w, conn, host, db, t, o, d, []Message{Message{Msg:sql2string(stmt), Rows: -1, Affected:affected, Seconds: sec}}, nextstmt)
	} else {
		shipMessage(w, host, db, "Where clauses not found")
	}
}


/* The next three functions deal with modifications in tables with primary key:
 * They use prepared statements.
 * However, these tempates only deal with values, not with identifiers. */

func actionKV_UPDATEFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, k string, v string) {
	hiddencols := []CContext{
		CContext{"", "k", "", "", "", "", "valid", k, ""},
		CContext{"", "v", "", "", "", "", "valid", v, ""}}
	stmt := sqlStar(t) + sqlWhere(k, "=", v)// TODO where1

	preparedStmt, err := sqlPrepare(conn, stmt)
	defer preparedStmt.Close()
	checkErrorPage(w, host, db, t, stmt, err)
	rows, err := preparedStmt.Query() //TODO query1
	checkY(err)
	defer rows.Close()
	primary := getPrimary(conn, t)
	shipForm(w, r, conn, host, db, t, "", "", "KV_UPDATE", "Update", "", getColumnInfoFilled(conn, host, db, t, primary, rows), hiddencols)
}

func actionKV_UPDATE(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, k string, v string) {
	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)
	_, sclauses, _ := collectClauses(r, conn, t)
	if len(sclauses) > 0 {
		stmt := sqlUpdate(t) + sqlSetClauses(sclauses) + sqlWhere(k, "=", v)// TODO where1
		preparedStmt, err := sqlPrepare(conn, stmt)
		defer preparedStmt.Close()
		checkErrorPage(w, host, db, t, stmt, err)
		result, sec, err := sqlExec(preparedStmt)// TODO exec1
		checkErrorPage(w, host, db, t, stmt, err)
		affected, err := result.RowsAffected()
		checkErrorPage(w, host, db, t, stmt, err)
		nextstmt := sqlStar(t)
		ms := []Message{Message{Msg:/*"?="+v + ": " + */sql2string(stmt), Rows: -1, Affected:affected, Seconds: sec}}
		dumpQuery(w, conn, host, db, t, "", "", ms, nextstmt)
	} else {
		shipMessage(w, host, db, "Set clauses not found")
	}
}

func actionKV_DELETE(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, k string, v string) {
	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)

	if k != "" && v != "" {
		stmt := sqlDelete(t) + sqlWhere(k, "=", v) // TODO where1
		preparedStmt, err := sqlPrepare(conn, stmt)
		defer preparedStmt.Close()
		checkErrorPage(w, host, db, t, stmt, err)
		result, sec, err := sqlExec(preparedStmt) // TODO exec1
		checkErrorPage(w, host, db, t, stmt, err)
		affected, err := result.RowsAffected()
		checkErrorPage(w, host, db, t, stmt, err)
		nextstmt := sqlStar(t)
		ms := []Message{Message{Msg:sql2string(stmt), Rows: -1, Affected:affected, Seconds: sec}}
		dumpQuery(w, conn, host, db, t, "", "", ms, nextstmt)
	} else {
		shipMessage(w, host, db, "Where clause not found")
	}

}

func actionGV_DELETE(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, g string, v string) {
	q := url.Values{}
	q.Add("db", db)
	q.Add("t", t)

	if g != "" && v != "" {
		stmt := sqlDelete(t) + sqlWhere(g, "=", v) // TODO where1
		preparedStmt, err := sqlPrepare(conn, stmt)
		defer preparedStmt.Close()
		checkErrorPage(w, host, db, t, stmt, err)
		result, sec, err := sqlExec(preparedStmt) // TODO exec1
		checkErrorPage(w, host, db, t, stmt, err)
		affected, err := result.RowsAffected()
		checkErrorPage(w, host, db, t, stmt, err)
		nextstmt := sqlStar(t)
		ms := []Message{Message{Msg:sql2string(stmt), Rows: -1, Affected:affected, Seconds: sec}}
		dumpQuery(w, conn, host, db, t, "", "", ms, nextstmt)
	} else {
		shipMessage(w, host, db, "Where clause not found")
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
