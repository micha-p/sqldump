package main

import (
	"database/sql"
	"log"
	"regexp"
	"strings"
)

/* sql escaping using prepared statements in Go is restricted, as it does not work for identifiers */

// special type to prevent mixture with normal strings
type sqlstring string

func sql2string(s sqlstring) string {
	return string(s)
}

func string2sql(s string) sqlstring {
	return sqlstring(s)
}

// Interface functions to underlying sql driver

func sqlPrepare(conn *sql.DB, s sqlstring) (*sql.Stmt, error) {
	stmt, err := conn.Prepare(sql2string(s))
	return stmt, err
}

func sqlQuery(conn *sql.DB, s sqlstring) (*sql.Rows, error) {
	stmtstr := string(s)
	log.Println("[SQL]", stmtstr)
	stmt, err := conn.Query(stmtstr)
	return stmt, err
}

func sqlExec(conn *sql.DB, s sqlstring) (*sql.Rows, error) {
	stmtstr := string(s)
	log.Println("[SQL]", stmtstr)
	stmt, err := conn.Query(stmtstr)
	return stmt, err
}


func sqlQueryRow(conn *sql.DB, s sqlstring) *sql.Row {
	return conn.QueryRow(sql2string(s))
}

// functions to prepare sql statements

func sqlStar(t string) sqlstring {
	return string2sql("SELECT * FROM ") + sqlProtectIdentifier(t)
}

func sqlSelect(c string, t string) sqlstring {
	return string2sql("SELECT ") + sqlProtectIdentifier(c) + string2sql(" FROM ") + sqlProtectIdentifier(t)
}

func sqlOrder(order string, desc string) sqlstring {
	var query sqlstring
	if order != "" {
		query = string2sql(" ORDER BY ") + sqlProtectIdentifier(order)
		if desc != "" {
			query = query + string2sql(" DESC")
		}
	}
	return query
}

// records start with number 1. Every child knows

func sqlLimit(limit int64, offset int64) sqlstring {
	query := string2sql(" LIMIT " + Int64toa(maxInt64(limit, 1)))
	if offset > 0 {
		query = query + string2sql(" OFFSET "+Int64toa(offset-1))
	}
	return query
}

func sqlCount(t string) sqlstring {
	return string2sql("SELECT COUNT(*) FROM ") + sqlProtectIdentifier(t)
}

func sqlColumns(t string) sqlstring {
	return string2sql("SHOW COLUMNS FROM ") + sqlProtectIdentifier(t)
}

func sqlInsert(t string) sqlstring {
	return string2sql("INSERT INTO ") + sqlProtectIdentifier(t)
}

func sqlUpdate(t string) sqlstring {
	return string2sql("UPDATE ") + sqlProtectIdentifier(t)
}

func sqlDelete(t string) sqlstring {
	return string2sql("DELETE FROM ") + sqlProtectIdentifier(t)
}

func sqlWhere(k string, c string, v string) sqlstring {
	if k =="" {
		return ""
	} else {
		return string2sql(" WHERE ") + sqlProtectIdentifier(k) + sqlFilterComparator(c) + sqlProtectString(v)
	}
}

// TODO: check usefulness for quick groups
func sqlHaving(g string, c string, v string) sqlstring {
	if g =="" {
		return ""
	} else {
		return string2sql(" HAVING ") + sqlProtectIdentifier(g) + sqlFilterComparator(c) + sqlProtectString(v)
	}
}

// from http://golang.org/src/strings/strings.go?h=Join#L382
func sqlJoin(a []sqlstring, sep string) sqlstring {
	if len(a) == 0 {
		return string2sql("")
	}
	if len(a) == 1 {
		return a[0]
	}
	n := len(sep) * (len(a) - 1)
	for i := 0; i < len(a); i++ {
		n += len(a[i])
	}

	b := make([]byte, n)
	bp := copy(b, a[0])
	for _, q := range a[1:] {
		s := sql2string(q)
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], s)
	}
	return sqlstring(b)
}

func sqlWhereClauses(clauses []sqlstring) sqlstring {
	if len(clauses) ==0 {
		return ""
	} else {
		return string2sql(" WHERE ") + sqlJoin(clauses, " AND ")
	}
}

func sqlSetClauses(clauses []sqlstring) sqlstring {
	if len(clauses) ==0 {
		return ""
	} else {
		return string2sql(" SET ") + sqlJoin(clauses, " , ")
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
		return string2sql("`" + r + "`")
	} else {
		return string2sql("`" + s + "`")
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
		return string2sql("\"" + r + "\"")
	} else {
		return string2sql("\"" + s + "\"")
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
	return string2sql("'" + re.FindString(t) + "'")
}

func sqlFilterComparator(t string) sqlstring {
	re := regexp.MustCompile("^ *(" + SQLCMP + ") *$")
	return string2sql(re.FindString(t))
}
