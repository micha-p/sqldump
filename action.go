package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"net/url"
	"sort"
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

func actionRouter(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string) {

	db, t, o, d, n, g, k, v := readRequest(r)

	q := r.URL.Query()
	action := q.Get("action")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	wclauses, sclauses, _ := collectClauses(r, conn, t)
	// TODO change *FORM to form="*"
	if action == "INFO" {
		stmt := str2sql("SHOW COLUMNS FROM ") + sqlProtectIdentifier(t)
		showInfo(w, conn, host, db, t, stmt)
	} else if action == "GOTO" && n != "" {
		dumpRouter(w, r, conn, host, db, t, o, d, n, g, k, v)
	} else if action == "SELECT" {
		dumpRouter(w, r, conn, host, db, t, o, d, n, g, k, v)
	} else if action == "BACK" {
		dumpRouter(w, r, conn, host, db, "", "", "", "", "", "", "")
	} else if action == "INSERT" && !READONLY && len(sclauses) > 0 {
		stmt := sqlInsert(t) + sqlSetClauses(sclauses)
		actionEXEC(w, conn, host, db, t, o, d, stmt)
	} else if action == "SELECTFORM" {
		actionSELECTFORM(w, r, conn, host, db, t, o, d)
	} else if action == "INSERTFORM" && !READONLY {
		actionINSERTFORM(w, r, conn, host, db, t, o, d)
	} else if action == "DELETEFORM" && !READONLY {
		actionDELETEFORM(w, r, conn, host, db, t, o, d)

	// TODO check Update insert and delete
	} else if action == "UPDATEFORM" && !READONLY && k != "" && v != "" {
		actionKV_UPDATEFORM(w, r, conn, host, db, t, k, v)
	} else if action == "UPDATEFORM" && !READONLY && g != "" && v != "" {
		actionGV_UPDATEFORM(w, r, conn, host, db, t, g, v)
	} else if action == "UPDATEFORM" && !READONLY {
		actionUPDATEFORM(w, r, conn, host, db, t, o, d)

	} else if action == "UPDATE" && !READONLY && k != "" && v != "" && len(sclauses) > 0 {
		hclause := sqlProtectIdentifier(k) + "=?"
		stmt := sqlUpdate(t) + sqlSetClauses(sclauses) + sqlWhereClauses(append(wclauses,hclause))
		actionEXEC1(w, conn, host, db, t, stmt, v)
	} else if action == "UPDATE" && !READONLY && g != "" && v != "" && len(sclauses) > 0 {
		hclause := sqlProtectIdentifier(g) + "=?"
		stmt := sqlUpdate(t) + sqlSetClauses(sclauses) + sqlWhereClauses(append(wclauses,hclause))
		actionEXEC1(w, conn, host, db, t, stmt, v)
	} else if action == "UPDATE" && !READONLY && len(sclauses) > 0 && len(wclauses) > 0 {
		stmt := sqlUpdate(t) + sqlSetClauses(sclauses) + sqlWhereClauses(wclauses)
		actionEXEC(w, conn, host, db, t, o, d, stmt)

	} else if action == "DELETE" && !READONLY && g != "" && v != "" {
		stmt := sqlDelete(t) + sqlWhere1(g, "=")
		actionEXEC1(w, conn, host, db, t, stmt, v)
	} else if action == "DELETE" && !READONLY && k != "" && v != "" {
		stmt := sqlDelete(t) + sqlWhere1(k, "=")
		actionEXEC1(w, conn, host, db, t, stmt, v)
	} else if action == "DELETE" && !READONLY && len(wclauses) > 0 {
		stmt := sqlDelete(t) + sqlWhereClauses(wclauses)
		actionEXEC(w, conn, host, db, t, o, d, stmt)

	} else {
		shipMessage(w, host, db, "Action unknown or insufficient parameters: "+action)
	}
}

// INSERTFORM and SELECTFORM provide columns without values, EDIT/UPDATE provide a filled vmap
// TODO: use DEFAULT and AUTOINCREMENT

func shipForm(w http.ResponseWriter, r *http.Request, conn *sql.DB,
	host string, db string, t string, o string, d string,
	action string, button string, selector string, showncols []CContext, hiddencols []CContext, whereStack []string) {

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
		Trail:    makeTrail(host, db, t, "", "", whereStack,url.Values{}),
	}

	if DEBUGFLAG {
		initTemplate()
	}
	err := templateForm.Execute(w, c)
	checkY(err)
}

/* The next four functions generate forms for doing SELECT, DELETE, INSERT, UPDATE */

func actionSELECTFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, o string, d string) {
	colinfo := getColumnInfo(conn, t)
	whereStack := WhereQuery2Pretty(r.URL.Query(), colinfo)
	shipForm(w, r, conn, host, db, t, o, d, "SELECT", "Select", "true", colinfo, []CContext{}, whereStack)
}

func actionDELETEFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, o string, d string) {
	colinfo := getColumnInfo(conn, t)
	whereStack := WhereQuery2Pretty(r.URL.Query(), colinfo)
	shipForm(w, r, conn, host, db, t, o, d, "DELETE", "Delete", "true", colinfo, []CContext{}, whereStack )
}

func actionINSERTFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, o string, d string) {
	colinfo := getColumnInfo(conn, t)
	whereStack := WhereQuery2Pretty(r.URL.Query(), colinfo)
	shipForm(w, r, conn, host, db, t, o, d, "INSERT", "Insert", "", colinfo, []CContext{}, whereStack)
}

// TODO combine next 3 to 1 function: always promote gk,v, always fill if count = 1
func actionUPDATEFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, o string, d string) {

	wclauses, _, whereQ := collectClauses(r, conn, t)
	colinfo := getColumnInfo(conn, t)
	whereStack := WhereQuery2Pretty(r.URL.Query(), colinfo)
	hiddencols := []CContext{}
	for field, valueArray := range whereQ { //type Values map[string][]string
		hiddencols = append(hiddencols, CContext{"", field, "", "", "", "", "valid", valueArray[0], ""})
	}

	count, _ := getSingleValue(conn, sqlCount(t)+sqlWhereClauses(wclauses))
	if count == "1" {
		rows, err, _ := getRows(conn, sqlStar(t)+sqlWhereClauses(wclauses))
		checkY(err)
		defer rows.Close()
		shipForm(w, r, conn, host, db, t, o, d, "UPDATE", "Update", "", getColumnInfoFilled(conn, host, db, t, "", rows), hiddencols, whereStack)
	} else {
		shipForm(w, r, conn, host, db, t, o, d, "UPDATE", "Update", "", colinfo, hiddencols, whereStack)
	}
}
func actionKV_UPDATEFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, k string, v string) {
	colinfo := getColumnInfo(conn, t)
	whereStack := WhereQuery2Pretty(r.URL.Query(), colinfo)
	isNumeric := colinfo[sort.Search(len(colinfo), func(i int) bool { return colinfo[i].Name == k })].IsNumeric
	whereStack = append(whereStack,whereComp2Pretty(k,"=",v,isNumeric))
	hiddencols := []CContext{
		CContext{"", "k", "", "", "", "", "valid", k, ""},
		CContext{"", "v", "", "", "", "", "valid", v, ""}}
	stmt := sqlStar(t) + sqlWhere1(k, "=")
	preparedStmt, _, err := sqlPrepare(conn, stmt)
	defer preparedStmt.Close()
	checkErrorPage(w, host, db, t, stmt, err)
	rows, _, err := sqlQuery1(preparedStmt, v)
	checkY(err)
	defer rows.Close()
	primary := getPrimary(conn, t)
	shipForm(w, r, conn, host, db, t, "", "", "UPDATE", "Update", "", getColumnInfoFilled(conn, host, db, t, primary, rows), hiddencols, whereStack)
}
func actionGV_UPDATEFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string, t string, g string, v string) {
	colinfo := getColumnInfo(conn, t)
	whereStack := WhereQuery2Pretty(r.URL.Query(), colinfo)
	isNumeric := colinfo[sort.Search(len(colinfo), func(i int) bool { return colinfo[i].Name == g })].IsNumeric
	whereStack = append(whereStack,whereComp2Pretty(g,"=",v,isNumeric))
	hiddencols := []CContext{
		CContext{"", "g", "", "", "", "", "valid", g, ""},
		CContext{"", "v", "", "", "", "", "valid", v, ""}}
	stmt := sqlStar(t) + sqlWhere1(g, "=")
	preparedStmt, _, err := sqlPrepare(conn, stmt)
	defer preparedStmt.Close()
	checkErrorPage(w, host, db, t, stmt, err)
	rows, _, err := sqlQuery1(preparedStmt, v)
	checkY(err)
	defer rows.Close()
	primary := getPrimary(conn, t)
	shipForm(w, r, conn, host, db, t, "", "", "UPDATE", "Update", "", getColumnInfoFilled(conn, host, db, t, primary, rows), hiddencols, whereStack)
}

// Excutes a statement on a selection by where-clauses
// Used, when rows are not adressable by a primary key or in table having a group
func actionEXEC(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, stmt sqlstring) {

	messageStack := []Message{}
	preparedStmt, sec, err := sqlPrepare(conn, stmt)
	defer preparedStmt.Close()
	checkErrorPage(w, host, db, t, stmt, err)
	messageStack = append(messageStack, Message{"PREPARE stmt FROM '" + sql2str(stmt) + "'", -1, 0, sec})

	result, sec, err := sqlExec(preparedStmt)
	checkErrorPage(w, host, db, t, stmt, err)
	affected, err := result.RowsAffected()
	checkErrorPage(w, host, db, t, stmt, err)

	messageStack = append(messageStack, Message{"EXECUTE stmt", -1, affected, sec})
	nextstmt := sqlStar(t) + sqlOrder(o, d)
	dumpRows(w, conn, host, db, t, o, d, nextstmt, messageStack)
}

/* Executes prepared statements about modifications in tables with primary key or having a group
 * Uses one argument as value for where clause
 * However, prepared statements only work with values, not in identifier position */
func actionEXEC1(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, stmt sqlstring, arg string) {

	messageStack := []Message{}
	preparedStmt, sec, err := sqlPrepare(conn, stmt)
	defer preparedStmt.Close()
	checkErrorPage(w, host, db, t, stmt, err)
	messageStack = append(messageStack, Message{"PREPARE stmt FROM '" + sql2str(stmt) + "'", -1, 0, sec})

	result, sec, err := sqlExec1(preparedStmt, arg)
	checkErrorPage(w, host, db, t, stmt, err)
	affected, err := result.RowsAffected()
	checkErrorPage(w, host, db, t, stmt, err)

	messageStack = append(messageStack, Message{"EXECUTE stmt USING \"" + arg + "\"", -1, affected, sec})
	nextstmt := sqlStar(t)
	dumpRows(w, conn, host, db, t, "", "", nextstmt, messageStack)
}
