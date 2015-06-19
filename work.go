package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net/url"
	"net/http"
	"html"
	"strings"
	"regexp"
)

/* dumpRows is the basic routine for any view without further restrictions. It starts with a fresh query.
 *
 *


//      ,------- home: 	lift restrictions from right lo left
//     /      ,------ indicator for primary key
//    /      /	   ,----- column:	ascending and descending order
//   #	 c1(ID)   c2	c3
//   1	   . 	   .	 .
//   2	   . \     . ----- value: select group  (browse by group)
//    \    .  `------ key value: show record  (browse by key values)
//     `----------- row number: show record  (browse by number or range)
//
//

//
//       __________  home: lift restrictions from right lo left
//      /   ___________ column: ascending and descending orde
//     /   /   ___________ indicator for primary key
//    /   /   /
//   #	 c1(ID)   c2	c3
//   1	   -       -     -
//   2	   - \     - \____ value: select group (browse by group)
//    \    -  \______ key value: show record (browse by key values)
//     \___________ row number: show record (browse by number or range)
//
//
//	View			Click				Query				Result			SQL
//	dumpRows
//		|------->	#		->			db t				dumpRows 		select	order
//		|------->	col		->			db t [od]			dumpRows 		select 							(horizontal)
//		|------->	2		->			db t [od] n=*		dumpFields 		select [order]	limit 1			(vertical)
//		|------->	1-3		->			db t [od] n=*-*		dumpRange 		select [order]	limit offset
//		|------->	_|_|	->			db t g v [od]		dumpGroup		select [order]	having g=v	 	(horizontal)
//		|------->	(ID)	->			db t k v 			dumpKeyValue    select 			having k=v	 	(vertical)
//		`------->	[?]		->FORM->	db t [od] w1...		dumpWhere      	select where [order]
//																|------->	#	 	->	dumpRows (ordered)
//																|------->	2	 	->	dumpFields
//																|------->	1-3	 	->	dumpRange
//																|------->	_|_|	->	dumpGroup
//																|------->	(ID)	->	dumpKeyValue
//																`------->	[?]		->	dumpWhere
//	dumpWhere
//		|------->	#		->			db t [od]			dumpRows		select where [order]
//		|------->	2		->			db t w [od] n=*		dumpFields		select where [order] limit 1
//		|------->	1-3		->			db t w [od] n=*-*	dumpRange		select where [order] limit offset
//		|------->	_|_|	->			db t g v [od]		dumpGroup		select where [order] having g=v
//		|------->	(ID)	->			db t [od] k v		dumpKeyValue	select where [order] having k=v
//		`------->	[?]		->FORM->	db t [od] w1 w2 ...	dumpWhere		select where w1 AND w2 [order]
//																|------->	#	 	->	dumpRows (ordered)
//																|------->	2	 	->	dumpFields
//																|------->	1-3	 	->	dumpRange
//																|------->	_|_|	->	dumpGroup
//																|------->	(ID)	->	dumpKeyValue
//																`------->	[?]		->	dumpWhere
//


///////////////////////		SQL:												Counter:		Menu
//
// Rows are not adressable:
//
// dumpRows  	    					    	[order]							empty			SELECTFORM	INSERTFORM	SELECTFORM	SELECTFORM	INFO
// dumpRange 				[where] 			[order] 	limit o-a+1, a-1,	range			SELECTFORM	INSERTFORM	SELECTFORM	SELECTFORM	INFO
// dumpField				[where] 			[order] 	limit 1,n-1			n				SELECTFORM	INSERTFORM	SELECTFORM	SELECTFORM	INFO
//
// Rows are selected by where-clause:
//
// dumpWhere	 			[where]		where	[order]							hidden			SELECTFORM	INSERTFORM	UPDATEFORM	DELETE		INFO
//
// Rows are adressed by key or groupvalue:
//
// dumpGroup				[where]    			[order]		having k=v			k=v				SELECTFORM	INSERTFORM	UPDATEFORM	DELETE		INFO
// dumpKeyValue				[where]  						having g=v 			g=v				SELECTFORM	INSERTFORM	UPDATEFORM	DELETE		INFO


/* showing always the same five menu entries introduces lesser changes in user interface.
 * Two subsequent forms might be confusing as well, on the other hand, insisting on select step might feel pedantic.
 */


func dumpRouter(w http.ResponseWriter, r *http.Request, conn *sql.DB,
	host string, db string, t string, o string, d string, n string, g string, k string, v string) {

	stmt := sqlStar(t)

    if k != "" && v != "" && k == getPrimary(conn, t) {
		stmt = stmt + sqlWhere(k, "=", v)
		dumpKeyValue(w, conn, host, db, t, k, v, stmt)
	} else {
		q := r.URL.Query()
		wclauses, _, _ := collectClauses(r, conn, t)

		if len(wclauses) > 0 {
			stmt = "SELECT TEMP.* FROM (" + stmt + sqlWhereClauses(wclauses) + ") TEMP "

			if g !="" && v !=""{
				stmt = stmt + sqlHaving(g, "=", v) + sqlOrder(o, d)
				dumpGroup(w, conn, host, db, t, o, d, g, v, stmt, q)
			} else {
				if o != "" {
					stmt = stmt + sqlOrder(o, d)
				}
				if n != "" {
					singlenumber := regexp.MustCompile("^ *(\\d+) *$").FindString(n)
					limits := regexp.MustCompile("^ *(\\d+) *- *(\\d+) *$").FindStringSubmatch(n)
					if singlenumber != "" {
						nint, _ := Atoi64(singlenumber)
						stmt = stmt + sqlLimit(2, nint) // for finding next record
						dumpFields(w, conn, host, db, t, o, d, singlenumber, nint, stmt, q)
					} else if len(limits) == 3 {
						startint, err := Atoi64(limits[1])
						checkY(err)
						endint, err := Atoi64(limits[2])
						checkY(err)
						maxint, err := Atoi64(getCount(conn, t))
						checkY(err)
						endint = minInt64(endint, maxint)
						stmt = stmt + sqlLimit(1+endint-startint, startint)
						dumpRange(w, conn, host, db, t, o, d, startint, endint, maxint, stmt, q)
					} else {
						shipMessage(w, host, db, "Can't convert to number or range: "+n)
					}
				} else {
					dumpWhere(w, conn, host, db, t, o, d, stmt, q)
				}
			}
		} else {
			stmt = stmt + sqlOrder(o, d)
			dumpRows(w, conn, host, db, t, o, d, stmt, []Message{})
		}
	}
}

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


func workRouter(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string) {

	db, t, o, d, n, g, k, v := readRequest(r)

	q := r.URL.Query()
	action := q.Get("action")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if action !=""  && db != "" && t != "" {
		actionRouter(w, r ,conn, host)
	} else if db == "" {
		showDatabases(w, conn, host)
	} else if t == "" {
		showTables(w, conn, host, db, t, o, d, g, v)
	} else {
		dumpRouter(w, r, conn, host, db, t, o, d, n, g, k, v)
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


