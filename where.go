package main

import (
	"net/http"
	"net/url"
	"strconv"
//	"fmt"
)

type opstring string

type Clause struct {
	Column    string
	Operator  opstring
	Value     string
	IsNumeric bool
}

/* Forms and Links provide where clauses encoded via two types of keys: Wncolumn and Oncolumn.
 * The first denotes the content of level n, the latter the operator of level n.
 *
 * They are submitted by forms as two-letter abbreviations and stored this way in type struct Clause.
 * Numeric clauses submitted without operator and parsed from textual input are stored have to be transofmed into two-lett abbreviations before stacking.
 *
 */

func clause2sql(c Clause) sqlstring {
	colname := sqlProtectIdentifier(c.Column)
	val := c.Value
	comp := c.Operator
	numeric := c.IsNumeric

	if comp == "s?" {
		return colname + op2sql(comp, numeric)
	} else if comp == "i0" {
		return colname + op2sql(comp, numeric)
	} else if comp == "n0" {
		return colname + op2sql(comp, numeric)
	} else if val != "" {
		if comp == "" {
			return colname + str2sql("=") + sqlProtectString(val) // default
		} else {
			if numeric {
				return colname + op2sql(comp, numeric) + sqlFilterNumber(val)
			} else {
				if comp == "eq" {
					return str2sql("BINARY ") + colname + op2sql(comp, numeric) + sqlProtectString(val)
				} else if comp == "ne" {
					return str2sql("BINARY ") + colname + op2sql(comp, numeric) + sqlProtectString(val)
				} else {
					return colname + op2sql(comp, numeric) + sqlProtectString(val)
				}
			}
		}
	} else {
		return ""
	}
}



func str2op(s string) opstring {
	n := make(map[string]opstring, 16) // numeric
	n[""] = "eq"
	n["="] = "eq"
	n["=="] = "eq"
	n["<>"] = "ne"
	n["!="] = "ne"
	n[">"] = "gt"
	n[">="] = "ge"
	n["<"] = "lt"
	n["<="] = "le"
	return n[s]
}

func op2sql(s opstring, numeric bool) sqlstring {
	return str2sql(op2str(s, numeric))
}

func op2str(s opstring, numeric bool) string {

	n := make(map[opstring]string, 16) // numeric
	n["eq"] = "="
	n["ne"] = "<>"
	n["gt"] = ">"
	n["ge"] = ">="
	n["lt"] = "<"
	n["le"] = "<="
	n["i0"] = " IS NULL"
	n["n0"] = " IS NOT NULL"
	n["sn"] = "="      // set number
	n["s0"] = "= NULL" // set NULL
	n["s?"] = "= ?"    // set ? using prepared statement

	m := make(map[opstring]string, 16) // strings
	m["lk"] = " LIKE "
	m["nl"] = " NOT LIKE "
	m["eq"] = "="  // equal binary -> case sensitive
	m["ne"] = "!=" // not equal binary
	m["gt"] = ">"
	m["ge"] = ">="
	m["lt"] = "<"
	m["le"] = "<="
	m["i0"] = " IS NULL"
	m["n0"] = " IS NOT NULL"
	m["sv"] = "="      // set value
	m["s0"] = "= NULL" // set NULL
	m["s?"] = "= ?"    // set ? using prepared statement

	var r string
	if numeric {
		r = n[s]
	} else {
		r = m[s]
	}
	return r
}

// TODO: return array of structs Clause;  use readerfunc
func collectClauses(r *http.Request, cols []CContext) ([][]Clause, []Clause) {

	var whereclauses [][]Clause
	var setclauses []Clause
	for i := 1; ; i++ {
		level := strconv.Itoa(i)
		wclausesOfLevel, sclausesOfLevel := collectClausesOfLevel(r, cols, level)
		if len(wclausesOfLevel) == 0 && len(sclausesOfLevel) == 0 {
			break
		} else {
			whereclauses = append(whereclauses, wclausesOfLevel)
			setclauses = append(setclauses, sclausesOfLevel...)
		}
	}
	return whereclauses, setclauses
}

func collectClausesOfLevel(r *http.Request, cols []CContext, level string) ([]Clause, []Clause) {

	var whereclauses, setclauses []Clause
	for _, col := range cols {
		colname := col.Name
		val := r.FormValue("W" + level + colname)
		set := r.FormValue("S" + level + colname)
		null := r.FormValue("N" + level + colname)
		comp := r.FormValue("O" + level + colname)
		var numeric bool
		if col.IsNumeric != "" {
			numeric = true
		} else {
			numeric = false
		}
		if null == "N" {
			setclauses = append(setclauses, Clause{Column: colname, Operator: "s0", IsNumeric: numeric})
		} else if null == "E" && !numeric {
			setclauses = append(setclauses, Clause{Column: colname, Operator: "sv", Value: "", IsNumeric: numeric})
		} else if set != "" && numeric {
			setclauses = append(setclauses, Clause{Column: colname, Operator: "sn", Value: set, IsNumeric: numeric})
		} else if set != "" && !numeric {
			setclauses = append(setclauses, Clause{Column: colname, Operator: "sv", Value: set, IsNumeric: numeric})
		}
		if val != "" || comp == "i0" || comp == "n0" {
			new := val2clause(colname, val, comp, col)
			whereclauses = append(whereclauses, new)
		}
	}
	return whereclauses, setclauses
}

func val2clause(colname string, val string, comp string, col CContext) Clause {
	var numeric bool
	if col.IsNumeric != "" {
		if comp == "" {
			ncomp, nval := sqlFilterNumericComparison(val)
			if ncomp == "" {
				comp = "eq" // use default for numeric comparisons
			} else {
				comp = string(str2op(ncomp))
				val = nval
			}
		}
		numeric = true
	} else {
		numeric = false
	}
	return Clause{colname, opstring(comp), val, numeric}
}

func WhereQuery2Stack(q url.Values, ccols []CContext) [][]Clause {

	var r [][]Clause
	for i := 1; ; i++ {
		level := strconv.Itoa(i)
		s := WhereQuery2Level(q, ccols, level)
		if len(s) == 0 {
			break
		} else {
			r = append(r, s)
		}
	}
	return r
}

func WhereQuery2Level(q url.Values, ccols []CContext, level string) []Clause {
	var clauses []Clause
	for _, col := range ccols {
		colname := col.Label
		val := q.Get("W" + level + col.Name)
		comp := q.Get("O" + level + col.Name)
		if val != "" || comp == "i0" || comp == "n0" {
			new := val2clause(colname, val, comp, col)
			clauses = append(clauses, new)
		}
	}
	return clauses
}

func putWhereStackIntoQuery(q url.Values, whereStack [][]Clause) {
	for i, whereClauses := range whereStack {
		level := strconv.Itoa(i + 1)
		putWhereClausesIntoQuery(q, level, whereClauses)
	}
}

func putWhereClausesIntoQuery(q url.Values, level string, whereClauses []Clause) {
	for _, clause := range whereClauses {
		colhtml := clause.Column
		if clause.Value != "" {
			q.Set("W"+level+colhtml, clause.Value)
		}
		q.Set("O"+level+colhtml, string(clause.Operator))
	}
}

func WhereStack2Hidden(whereStack [][]Clause) []CContext {
	var r []CContext
	for i, whereClauses := range whereStack {
		level := strconv.Itoa(i + 1)
		r = append(r, WhereClauses2Hidden(level, whereClauses)...)
	}
	return r
}

func WhereClauses2Hidden(level string, whereClauses []Clause) []CContext {
	var r []CContext
	for _, clause := range whereClauses {
		colhtml := clause.Column
		if clause.Value != "" {
			r = append(r, CContext{Name: "W" + level + colhtml, Value: clause.Value})
		}
		r = append(r, CContext{Name: "O" + level + colhtml, Value: string(clause.Operator)})
	}
	return r
}

func whereClauses2Pretty(whereClauses []Clause) string {
	var r string
	for i, clause := range whereClauses {
		if i > 0 {
			r = r + ", "
		}
		if clause.IsNumeric {
			r = r + clause.Column + op2str(clause.Operator, clause.IsNumeric) + clause.Value
		} else {
			if clause.Operator == "i0" {
				r = r + clause.Column + " IS NULL"
			} else if clause.Operator == "n0" {
				r = r + clause.Column + " IS NOT NULL"
			} else if clause.Operator == "eq" {
				r = r + clause.Column + "==" + "\"" + clause.Value + "\""
			} else {
				r = r + clause.Column + op2str(clause.Operator, clause.IsNumeric) + "\"" + clause.Value + "\""
			}
		}
	}
	return r
}
