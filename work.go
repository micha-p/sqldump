package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net/url"
	"net/http"
	"html"
	"strings"
)

// Output:
// rows 			dump.go -> dumpRows, dumpRange, dumpWhere, dumpGroup, dumpWhereGrup
// fields 			fields.go -> dumpFields dumpKeyValue
// simple table		toplevel.go -> dumpHome dumpTables dumpInfo


// rows are not adressable:
// dumpRows  -> SELECTFORM, INSERTFORM, PLEASESELECT, PLEASESELECT, INFO
// dumpRange -> SELECTFORM, INSERTFORM, PLEASESELECT, PLEASESELECT, INFO
// dumpField -> SELECTFORM, INSERTFORM, PLEASESELECT, PLEASESELECT, INFO
//
// rows are selected by where-clause:
// dumpWhere 		-> SELECTFORM, INSERTFORM, UPDATEFORM, DELETE, INFO
//
// rows are selected by key or group:
// dumpKeyValue 	-> SELECTFORM, INSERTFORM, UPDATEFORM, DELETE, INFO
// dumpGroup	 	-> SELECTFORM, INSERTFORM, UPDATEFORM, DELETE, INFO
//
// rows are selected by where and further identified by having group value
// dumpWhereGrouped	-> SELECTFORM, INSERTFORM, UPDATEFORM, DELETE, INFO
//
// toplevel
// dumpHome			-> depreciated (ACTION USEFORM)
// dumpTables		-> INSERTFORM, ALTERFORM, DROP, INFODB


/* showing always the same five menu entries introduces lesser changes in user interface.
 * Two subsequent forms might be confusing as well, on the other hand, insisting on select step might feel pedantic.
 */

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
		actionRouter(w, r ,conn, host)
	} else {
		dumpIt(w, r, conn, host, db, t, o, d, n, g, k, v)
	}
}

// TODO: to allow for submitting multiple clauses for a field, they should be numbered W1, O1 ...

// do not export
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

