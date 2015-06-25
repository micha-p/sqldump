package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"net/url"
	"regexp"
	//	"fmt"
)

/* dumpRows is the basic routine for any view without further restrictions. It starts with a fresh query.
 *
 *


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
	t string, o string, d string, n string, g string, k string, v string) {

	colinfo := getColumnInfo(conn, t)
	stmt := sqlStar(t)

	if k != "" && v != "" && k == getPrimary(conn, t) {
		stmt = stmt + sqlHaving(k, "=", v)
		showKeyValue(w, conn, t, o, d, k, v, stmt)
	} else {
		wclauses, _ := collectClauses(r, colinfo)

		if len(wclauses) == 0 && g == "" && v == "" && n == "" {
			stmt = stmt + sqlOrder(o, d)
			dumpRows(w, conn, t, o, d, stmt, []Message{})

		} else {
			stmt = stmt + sqlWhereClauses(wclauses)
			stmt = stmt + sqlHaving(g, "=", v)
			stmt = stmt + sqlOrder(o, d)
			dumpSelection(w, conn, t, o, d, n, g, v, stmt, wclauses, []Message{})
		}
	}
}

func dumpSelection(w http.ResponseWriter, conn *sql.DB, t string, o string, d string, n string, g string, v string,
	stmt sqlstring, whereStack [][]Clause, messageStack []Message) {

	if g == "" && n == "" {
		dumpWhere(w, conn, t, o, d, stmt, whereStack, messageStack)
	} else {
		if n == "" {
			dumpGroup(w, conn, t, o, d, g, v, stmt, whereStack, messageStack)
		} else {
			singlenumber := regexp.MustCompile("^ *(\\d+) *$").FindString(n)
			limits := regexp.MustCompile("^ *(\\d+) *- *(\\d+) *$").FindStringSubmatch(n)
			nmax, err := Atoi64(getCount(conn, sqlCount(t)+sqlWhereClauses(whereStack)))
			checkY(err)
			if singlenumber != "" {
				ni, _ := Atoi64(singlenumber)
				ni = minInt64(ni, nmax)
				stmt = stmt + sqlLimit(1, ni)
				showFields(w, conn, t, o, d, singlenumber, ni, nmax, stmt, whereStack)
			} else if len(limits) == 3 {
				nstart, err := Atoi64(limits[1])
				checkY(err)
				nend, err := Atoi64(limits[2])
				checkY(err)
				nend = minInt64(nend, nmax)
				stmt = stmt + sqlLimit(1+nend-nstart, nstart)
				dumpRange(w, conn, t, o, d, nstart, nend, nmax, stmt, whereStack, messageStack)
			} else {
				shipMessage(w, conn, "Can't understand number or range: "+n)
			}
		}
	}
}



func readRequest(r *http.Request) (string, string, string, string, string, string, string) {
	q := r.URL.Query()
	t := q.Get("t")
	o := q.Get("o")
	d := q.Get("d")
	n := q.Get("n")
	g := q.Get("g")
	k := q.Get("k")
	v := q.Get("v")
	return t, o, d, n, g, k, v
}




func makeMenu(q url.Values, name string, value string, label string) Entry {
	return makeMenuPath(q,name, value, label,"")
}

func makeMenuPath(q url.Values, name string, value string, label string, path string) Entry {
	if name != "" {
		q.Set(name, value)
	}
	link := q.Encode()
	q.Del(name)
	return escape(label, path, link)
}

func makeMenu5(m url.Values) []Entry {
	var menu []Entry
	menu = append(menu, makeMenu(m, "action", "SELECTFORM", "?"))
	menu = append(menu, makeMenu(m, "action", "INSERTFORM", "+"))
	menu = append(menu, makeMenu(m, "action", "UPDATEFORM", "~")) // KV-DELETE, GV-DELETE
	menu = append(menu, makeMenu(m, "action", "DELETE", "-"))     // DELETEFILLED, KV-DELETE, GV-DELETE
	menu = append(menu, makeMenu(m, "action", "INFO", "i"))
	return menu
}

func makeMenu3(m url.Values) []Entry {
	var menu []Entry
	menu = append(menu, makeMenu(m, "action", "SELECTFORM", "?"))
	menu = append(menu, makeMenu(m, "action", "INSERTFORM", "+"))
	menu = append(menu, makeMenu(m, "", "", " "))
	menu = append(menu, makeMenu(m, "", "", " "))
	menu = append(menu, makeMenu(m, "action", "INFO", "i"))
	return menu
}

