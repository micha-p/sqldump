package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"html"
	"net/http"
	"net/url"
	"strconv"
)

// TODO: return array of structs Clause; remove whereQ; use readerfunc

func collectClauses(r *http.Request, conn *sql.DB, t string) ([][]sqlstring, []sqlstring, url.Values) {

	cols := getCols(conn, t)
	whereQ := url.Values{}
	var whereclauses [][]sqlstring
	var setclauses []sqlstring
	for i := 1;; i++ {
		level := strconv.Itoa(i)
		wclausesOfLevel,sclausesOfLevel,_ := collectClausesOfLevel(r, cols,level, whereQ)
		if len(wclausesOfLevel)==0 && len(sclausesOfLevel)==0 {
			break
		} else {
			whereclauses=append(whereclauses,wclausesOfLevel)
			setclauses=append(setclauses,sclausesOfLevel...)
		}
	}
	return whereclauses, setclauses, whereQ
}

// TODO: remove whereQ

func collectClausesOfLevel(r *http.Request, cols []string, level string, whereQ url.Values) ([]sqlstring, []sqlstring, url.Values) {

	var whereclauses, setclauses []sqlstring
	for _, col := range cols {
		colname := sqlProtectIdentifier(col)
		colhtml := html.EscapeString(col)
		val := r.FormValue("W"+level+col)
		set := r.FormValue("S"+level+col)
		null := r.FormValue("N"+level+col)
		comp := r.FormValue("O"+level+col)
		if val != "" || comp == "=0" || comp == "!0" {
			whereQ.Add("W"+level+colhtml, val)
			if comp == "" {
				comp, val = sqlFilterNumericComparison(val)
				whereclauses = append(whereclauses, colname+sqlFilterComparator(comp)+sqlFilterNumber(val))
			} else if comp == "=" {
				whereQ.Add("W"+level+colhtml, comp)
				whereclauses = append(whereclauses, colname+str2sql(" = ")+sqlProtectString(val))
			} else if comp == "~" {
				whereQ.Add("W"+level+colhtml, comp)
				whereclauses = append(whereclauses, colname+str2sql(" LIKE ")+sqlProtectString(val))
			} else if comp == "!~" {
				whereQ.Add("W"+level+colhtml, comp)
				whereclauses = append(whereclauses, colname+str2sql(" NOT LIKE ")+sqlProtectString(val))
			} else if comp == "==" {
				whereQ.Add("W"+level+colhtml, comp)
				whereclauses = append(whereclauses, str2sql("BINARY ")+colname+str2sql("=")+sqlProtectString(val))
			} else if comp == "!=" {
				whereQ.Add("W"+level+colhtml, comp)
				whereclauses = append(whereclauses, str2sql("BINARY ")+colname+str2sql("!=")+sqlProtectString(val))
			} else if comp == "=0" {
				whereQ.Add("W"+level+colhtml, comp)
				whereclauses = append(whereclauses, colname+str2sql(" IS NULL"))
			} else if comp == "!0" {
				whereQ.Add("W"+level+colhtml, comp)
				whereclauses = append(whereclauses, colname+str2sql(" IS NOT NULL"))
			} else {
				whereQ.Add("W"+level+colhtml, comp)
				if sqlFilterNumber(val) != "" {
					whereclauses = append(whereclauses, colname+sqlFilterComparator(comp)+sqlFilterNumber(val))
				} else {
					whereclauses = append(whereclauses, colname+sqlFilterComparator(comp)+sqlProtectString(val))
				}
			}
		}
		if null == "N" {
			whereQ.Add("N"+level+colhtml, null)
			setclauses = append(setclauses, colname+"=NULL")
		} else if null == "E" {
			whereQ.Add("N"+level+colhtml, null)
			setclauses = append(setclauses, colname+"=\"\"")
		} else if set != "" {
			whereQ.Add("S"+level+colhtml, set)
			setclauses = append(setclauses, colname+"="+sqlProtectString(set))
		} else if set != "" {
			whereQ.Add("S"+level+colhtml, set)
			setclauses = append(setclauses, colname+"="+sqlProtectString(set))
		}
	}
	return whereclauses, setclauses, whereQ
}


func WhereQuery2Stack(q url.Values, ccols []CContext) [][]Clause {

	var r [][]Clause
	for i := 1;; i++ {
		level := strconv.Itoa(i)
		s:= WhereQuery2Level(q, ccols, level)
		if len(s)==0 {
			break
		} else {
			r=append(r,s)
		}
	}
	return r
}



func WhereQuery2Level(q url.Values, ccols []CContext, level string) []Clause {
	var clauses []Clause
	for _, col := range ccols {
		colname := col.Label
		val := q.Get(html.EscapeString("W"+level+col.Name))
		comp := q.Get(html.EscapeString("O"+level+col.Name))
		if comp==""{
			comp, val = sqlFilterNumericComparison(html.UnescapeString(val))
		}
		if val != "" || comp == "=0" || comp == "!0" {
			var numeric bool

			if col.IsNumeric !="" {
				numeric = true
			} else {
				numeric = false
			}
			clauses = append(clauses,Clause{colname, comp, val, numeric}) // TODO check security: strings or sqlstrings?
		}
	}
	return clauses
}

func whereComp2Pretty(colname string, comp string, val string, IsNumeric string) string {
	var r string
	if comp == "" {
		comp, val = sqlFilterNumericComparison(val)
		r = colname+sql2str(sqlFilterComparator(comp))+sql2str(sqlFilterNumber(val))
	} else if comp == "~" {
		r = colname+" LIKE \""+val+"\""
	} else if comp == "!~" {
		r = colname+" NOT LIKE \""+val+"\""
	} else if comp == "==" {
		r = colname+"==\""+val+"\""
	} else if comp == "!=" {
		r = colname+"!=\""+val+"\""
	} else if comp == "=0" {
		r = colname+" IS NULL"
	} else if comp == "!0" {
		r = colname+" IS NOT NULL"
	} else {
		if IsNumeric == "" {
			r = colname+sql2str(sqlFilterComparator(comp))+" \""+val+"\""
		} else {
			r = colname+sql2str(sqlFilterComparator(comp))+sql2str(sqlFilterNumber(val))
		}
	}
	return r
}

func WhereQuery2Hidden1Level(q url.Values, ccols []CContext, level string) []CContext {
	var clauses []CContext
	for _, col := range ccols {
		val := q.Get(html.EscapeString("W"+level+col.Name))
		comp := q.Get(html.EscapeString("O"+level+col.Name))
		if val != "" {
			clauses = append(clauses,CContext{Name: "W"+level+col.Name, Value: val})
			clauses = append(clauses,CContext{Name: "O"+level+col.Name, Value: comp})
		}
		if comp == "=0" || comp == "!0" {
			clauses = append(clauses,CContext{Name: "O"+level+col.Name, Value: comp})
		}
	}
	return clauses
}

func WhereQuery2Hidden(q url.Values, ccols []CContext) []CContext {
	var r []CContext
	for i := 1;; i++ {
		level := strconv.Itoa(i)
		rl:= WhereQuery2Hidden1Level(q, ccols, level)
		if len(rl)==0 {
			break
		} else {
			r=append(r,rl...)
		}
	}
	return r
}

func whereComp2Hidden(colname string, comp string, val string, IsNumeric string) string {
	var r string
	if comp == "" {
		comp, val = sqlFilterNumericComparison(val)
		r = colname+sql2str(sqlFilterComparator(comp))+sql2str(sqlFilterNumber(val))
	} else if comp == "~" {
		r = colname+" LIKE \""+val+"\""
	} else if comp == "!~" {
		r = colname+" NOT LIKE \""+val+"\""
	} else if comp == "==" {
		r = colname+"==\""+val+"\""
	} else if comp == "!=" {
		r = colname+"!=\""+val+"\""
	} else if comp == "=0" {
		r = colname+" IS NULL"
	} else if comp == "!0" {
		r = colname+" IS NOT NULL"
	} else {
		if IsNumeric == "" {
			r = colname+sql2str(sqlFilterComparator(comp))+" \""+val+"\""
		} else {
			r = colname+sql2str(sqlFilterComparator(comp))+sql2str(sqlFilterNumber(val))
		}
	}
	return r
}

type Clause struct{
	Column string
	Comparator string
	Value string
	IsNumeric bool
}

func putWhereStackIntoQuery(q url.Values,whereStack [][]Clause) {
	for i,whereClauses := range(whereStack){
		level := strconv.Itoa(i+1)
		putWhereClausesIntoQuery(q,level,whereClauses)
	}
}

func putWhereClausesIntoQuery(q url.Values,level string, whereClauses []Clause) {
	for i,clause := range(whereClauses){
		level := strconv.Itoa(i+1)
		colhtml := html.EscapeString(clause.Column)
		if clause.Value != "" {
			q.Set("W"+level+colhtml,html.EscapeString(clause.Value))
		}
		q.Set("O"+level+colhtml,html.EscapeString(clause.Comparator))
	}
}

func whereClauses2Pretty(whereClauses []Clause) string{
	var r string
	for i,clause :=range(whereClauses){
		if i>0 {
			r=r+", "
		}
		if clause.Comparator == "~" {
			r = r + clause.Column+" LIKE \""+clause.Value+"\""
		} else if clause.Comparator == "!~" {
			r = r + clause.Column+" NOT LIKE \""+clause.Value+"\""
		} else if clause.Comparator == "==" {
			r = r + clause.Column+"==\""+clause.Value+"\""
		} else if clause.Comparator == "!=" {
			r = r + clause.Column+"!=\""+clause.Value+"\""
		} else if clause.Comparator == "=0" {
			r = r + clause.Column+" IS NULL"
		} else if clause.Comparator == "!0" {
			r = r + clause.Column+" IS NOT NULL"
		} else {
			if clause.IsNumeric {
				r = r + clause.Column+clause.Comparator+clause.Value
			} else {
				r = r + clause.Column+clause.Comparator+"\""+clause.Value+"\""
			}
		}
	}
	return r
}




