package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"net/url"
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

	wclauses, sclauses, whereQ := collectClauses(r, conn, t)
	// TODO change *FORM to form="*"
	if action == "INFO" {
		stmt := string2sql("SHOW COLUMNS FROM ") + sqlProtectIdentifier(t)
		dumpInfo(w, conn, host, db, t, stmt)
	} else if action == "GOTO" && n != "" {
		dumpIt(w, r, conn, host, db, t, o, d, n, g, k, v)
	} else if action == "BACK" {
		dumpIt(w, r, conn, host, db, "", "", "", "", "", "", "")
	} else if action == "SELECTFORM" {
		actionSELECTFORM(w, r, conn, host, db, t, o, d)
	} else if action == "INSERTFORM" && !READONLY {
		actionINSERTFORM(w, r, conn, host, db, t, o, d)
	} else if action == "DELETEFORM" && !READONLY { // Create subset for DELETE
		actionDELETEFORM(w, r, conn, host, db, t, o, d)
	} else if action == "UPDATEFORM" && !READONLY { // ask for changed values
		actionUPDATEFORM(w, r, conn, host, db, t, o, d)
	} else if action == "KV_UPDATEFORM" && !READONLY && k != "" && v != "" {
		actionKV_UPDATEFORM(w, r, conn, host, db, t, k, v)
	} else if action == "GV_UPDATEFORM" && !READONLY && g != "" && v != "" {
		actionGV_UPDATEFORM(w, r, conn, host, db, t, g, v)
	} else if action == "SELECT"  && len(wclauses)> 0 {
		stmt := sqlStar(t) + sqlWhereClauses(wclauses)
		// actionEXEC(w, conn, host, db, t, o, d, stmt)
		dumpWhere(w, conn, host, db, t, o, d, stmt, whereQ)
	} else if action == "INSERT" && !READONLY  && len(sclauses)> 0 {
		stmt := sqlInsert(t) + sqlSetClauses(sclauses)
		actionEXEC(w, conn, host, db, t, o, d, stmt)
	} else if action == "DELETE" && !READONLY  && len(wclauses)> 0 {
		stmt := sqlDelete(t) + sqlWhereClauses(wclauses)
		actionEXEC(w, conn, host, db, t, o, d, stmt)
	} else if action == "UPDATE" && !READONLY  && len(sclauses)> 0  && len(wclauses)> 0 {
		stmt := sqlUpdate(t) + sqlSetClauses(sclauses) + sqlWhereClauses(wclauses)
		actionEXEC(w, conn, host, db, t, o, d, stmt)

	} else if action == "KV_UPDATE" && !READONLY && k != "" && v != ""  && len(sclauses)> 0 {
		stmt := sqlUpdate(t) + sqlSetClauses(sclauses) + sqlWhere1(k, "=")
		actionEXEC1(w, conn, host, db, t, stmt, v)
	} else if action == "KV_DELETE" && !READONLY && k != "" && v != "" {
		stmt := sqlDelete(t) + sqlWhere1(k, "=")
		actionEXEC1(w, conn, host, db, t, stmt, v)
	} else if action == "GV_UPDATE" && !READONLY && g != "" && v != ""  && len(sclauses)> 0{
		stmt := sqlUpdate(t) + sqlSetClauses(sclauses) + sqlWhere1(g, "=")
		actionEXEC1(w, conn, host, db, t, stmt, v)
	} else if action == "GV_DELETE" && !READONLY && g != "" && v != "" {
		stmt := sqlDelete(t) + sqlWhere1(g, "=")
		actionEXEC1(w, conn, host, db, t, stmt, v)
	} else {
		shipMessage(w, host, db, "Action unknown or insufficient parameters: "+action)
	}
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



