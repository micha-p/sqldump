package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net/url"
	"net/http"
	"html"
	"strings"
)


// rows are not adressable:
// dumpRows  -> SELECTFORM, INSERTFORM, INFO
// dumpRange -> SELECTFORM, INSERTFORM, INFO
// dumpField -> SELECTFORM, INSERTFORM, INFO

// rows are selected by where-clause
// dumpWhere 		-> SELECTFORM, INSERTFORM, UPDATEFORM, DELETE, INFO

// rows are selected by key or group
// dumpKeyValue 	-> SELECTFORM, INSERTFORM, UPDATEFORM, DELETE, INFO
// dumpGroup	 	-> SELECTFORM, INSERTFORM, UPDATEFORM, DELETE, INFO


func readRequest(r *http.Request) (string, string, string, string, string, string, string, string) {
	q := r.URL.Query()
	db := q.Get("db")
	t := q.Get("t")
	o := q.Get("o")
	d := q.Get("d")
	n := q.Get("n")
	g := q.Get("g")
	k := q.Get("k")
	v := q.Get("v")
	return db, t, o, d, n, g, k, v
}

func workload(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string) {

	db, t, o, d, n, g, k, v := readRequest(r)

	q := r.URL.Query()
	action := q.Get("action")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if action !=""  && db != "" && t != "" {
		wclauses, sclauses, whereQ := collectClauses(r, conn, t)
		// TODO change *FORM to form="*"
		if action == "INFO" {
			actionINFO(w, r, conn, host, db, t)
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

		} else if action == "GOTO" && n != "" {
			dumpIt(w, r, conn, host, db, t, o, d, n, g, k, v)
		} else if action == "BACK" {
			dumpIt(w, r, conn, host, db, "", "", "", "", "", "", "")
		} else {
			shipMessage(w, host, db, "Action unknown or insufficient parameters: "+action)
		}
	} else {
		dumpIt(w, r, conn, host, db, t, o, d, n, g, k, v)
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
