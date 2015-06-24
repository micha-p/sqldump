package main

import (
	"database/sql"
	"log"
	"regexp"
	"strings"
	"time"
)

/* sql escaping using prepared statements is restricted, as it does not work for identifiers */

// special type to prevent mixture with normal strings
type sqlstring string

func sql2str(s sqlstring) string {
	return string(s)
}

func str2sql(s string) sqlstring {
	return sqlstring(s)
}

// Interface functions to underlying sql driver

func sqlPrepare(conn *sql.DB, s sqlstring) (*sql.Stmt, float64, error) {
	t0 := time.Now()
	stmtstr := sql2str(s)
	log.Println("[SQL]", stmtstr)
	r, err := conn.Prepare(stmtstr)
	t1 := time.Now()
	return r, t1.Sub(t0).Seconds(), err
}

func sqlQueryInternal(conn *sql.DB, s sqlstring) (*sql.Rows, float64, error) {
	stmtstr := string(s)
	stmt, err := conn.Query(stmtstr)
	return stmt, 0, err
}

func sqlQuery(conn *sql.DB, s sqlstring) (*sql.Rows, float64, error) {
	stmtstr := string(s)
	log.Println("[SQL]", stmtstr)
	t0 := time.Now()
	stmt, err := conn.Query(stmtstr)
	t1 := time.Now()
	return stmt, t1.Sub(t0).Seconds(), err
}

func sqlQuery1(prepared *sql.Stmt, arg string) (*sql.Rows, float64, error) {
	t0 := time.Now()
	r, err := prepared.Query(arg)
	t1 := time.Now()
	return r, t1.Sub(t0).Seconds(), err
}

func sqlExec(prepared *sql.Stmt) (sql.Result, float64, error) {
	t0 := time.Now()
	r, err := prepared.Exec()
	t1 := time.Now()
	return r, t1.Sub(t0).Seconds(), err
}

func sqlExec1(prepared *sql.Stmt, arg string) (sql.Result, float64, error) {
	log.Println("[SQL]", arg)
	t0 := time.Now()
	r, err := prepared.Exec(arg)
	t1 := time.Now()
	return r, t1.Sub(t0).Seconds(), err
}

func sqlQueryRow(conn *sql.DB, s sqlstring) *sql.Row {
	return conn.QueryRow(sql2str(s))
}

// functions to prepare sql statements

func sqlStar(t string) sqlstring {
	return str2sql("SELECT * FROM ") + sqlProtectIdentifier(t)
}

func sqlSelect(c string, t string) sqlstring {
	return str2sql("SELECT ") + sqlProtectIdentifier(c) + str2sql(" FROM ") + sqlProtectIdentifier(t)
}

func sqlOrder(order string, desc string) sqlstring {
	var query sqlstring
	if order != "" {
		query = str2sql(" ORDER BY ") + sqlProtectIdentifier(order)
		if desc != "" {
			query = query + str2sql(" DESC")
		}
	}
	return query
}

// records start with number 1. Every child knows
func sqlLimit(limit int64, offset int64) sqlstring {
	var query sqlstring
	if limit >= 0 {
		query = str2sql(" LIMIT " + Int64toa(maxInt64(limit, 1)))
		if offset > 0 {
			query = query + str2sql(" OFFSET "+Int64toa(offset-1))
		}
	}
	return query
}

func sqlCount(t string) sqlstring {
	return str2sql("SELECT COUNT(*) FROM ") + sqlProtectIdentifier(t)
}

func sqlColumns(t string) sqlstring {
	return str2sql("SHOW COLUMNS FROM ") + sqlProtectIdentifier(t)
}

func sqlInsert(t string) sqlstring {
	return str2sql("INSERT INTO ") + sqlProtectIdentifier(t)
}

func sqlUpdate(t string) sqlstring {
	return str2sql("UPDATE ") + sqlProtectIdentifier(t)
}

func sqlDelete(t string) sqlstring {
	return str2sql("DELETE FROM ") + sqlProtectIdentifier(t)
}

func sqlWhere(k string, c string, v string) sqlstring {
	if k == "" {
		return ""
	} else {
		return str2sql(" WHERE ") + sqlProtectIdentifier(k) + sqlFilterComparator(c) + sqlProtectString(v)
	}
}

func sqlWhere1(k string, c string) sqlstring {
	if k == "" {
		return ""
	} else {
		return str2sql(" WHERE ") + sqlProtectIdentifier(k) + sqlFilterComparator(c) + "?"
	}
}

// TODO: check usefulness for quick groups
func sqlHaving(g string, c string, v string) sqlstring {
	if g == "" {
		return ""
	} else {
		return str2sql(" HAVING ") + sqlProtectIdentifier(g) + sqlFilterComparator(c) + sqlProtectString(v)
	}
}

func sqlWhereClauses(whereStack [][]Clause) sqlstring {
	if len(whereStack) == 0 {
		return ""
	} else {
		var r sqlstring
	 	for _,clauses := range(whereStack) {
			for _,clause := range(clauses) {
				if len(r) > 0 {
					r = r +  " && "
				}
				r = r +  clause2sql(clause)
			}
		}
	return str2sql(" WHERE ") + r
	}
}

func sqlSetClauses(clauses []Clause) sqlstring {
	if len(clauses) == 0 {
		return ""
	} else {
		var r sqlstring
		for _,clause := range(clauses) {
			if len(r) > 0 {
				r = r +  ", "
			}
			r = r +  clause2sql(clause)
		}
	return str2sql(" SET ") + r
	}
}

// Filter and Escapes

/* 	https://dev.mysql.com/doc/refman/5.1/en/identifiers.html

    Identifiers are converted to Unicode internally. They may contain these characters:
    Permitted characters in unquoted identifiers:
        ASCII: [0-9,a-z,A-Z$_] (basic Latin letters, digits 0-9, dollar, underscore)
        Extended: U+0080 .. U+FFFF

    Permitted characters in quoted identifiers include the full Unicode Basic Multilingual Plane (BMP), except U+0000:
        ASCII: U+0001 .. U+007F
        Extended: U+0080 .. U+FFFF

    ASCII NUL (U+0000) and supplementary characters (U+10000 and higher) are not permitted in quoted or unquoted identifiers.
    Identifiers may begin with a digit but unless quoted may not consist solely of digits.
    Database, table, and column names cannot end with space characters.
    Before MySQL 5.1.6, database and table names cannot contain “/”, “\”, “.”, or characters that are not permitted in file names.
	The identifier quote character is the backtick (“`”):
*/

func sqlProtectIdentifier(s string) sqlstring {
	if s != "" && strings.ContainsAny(s, "`") {
		r := strings.Replace(s, "`", "``", -1)
		if DEBUGFLAG {
			log.Println("[SQLINJECTION?]", s+" -> "+r)
		}
		return str2sql("`" + r + "`")
	} else {
		return str2sql("`" + s + "`")
	}
}

/* 	https://dev.mysql.com/doc/refman/5.7/en/string-literals.html

	A string is a sequence of bytes or characters, enclosed within either single quote (“'”) or double quote (“"”) characters.

	There are several ways to include quote characters within a string:
    A “'” inside a string quoted with “'” may be written as “''”.
    A “"” inside a string quoted with “"” may be written as “""”.
    Precede the quote character by an escape character (“\”).
    A “'” inside a string quoted with “"” needs no special treatment and need not be doubled or escaped.
    In the same way, “"” inside a string quoted with “'” needs no special treatment.
*/

func sqlProtectString(s string) sqlstring {
	if s != "" && strings.ContainsAny(s, "\"") {
		r := strings.Replace(s, "\"", "\"\"", -1)
		if DEBUGFLAG {
			log.Println("[SQLINJECTION?]", s+" -> "+r)
		}
		return str2sql("\"" + r + "\"")
	} else {
		return str2sql("\"" + s + "\"")
	}
}

var SQLCMP = "[<=>]+"
var SQLNUM = "-?[0-9]+.?[0-9]*(E-?[0-9]+)?"

func sqlFilterNumericComparison(t string) (string, string) {
	re := regexp.MustCompile("^ *(" + SQLCMP + ") *(" + SQLNUM + ") *$")
	rm := re.FindStringSubmatch(t)
	if len(rm) > 2 {
		return rm[1], rm[2]
	} else {
		return "", ""
	}
}

func sqlFilterNumber(t string) sqlstring {
	re := regexp.MustCompile("^ *(" + SQLNUM + ") *$")
	return str2sql("'" + re.FindString(t) + "'")
}

func sqlFilterComparator(t string) sqlstring {
	re := regexp.MustCompile("^ *(" + SQLCMP + ") *$")
	return str2sql(re.FindString(t))
}
