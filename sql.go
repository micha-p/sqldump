package main

import (
	"log"
	"regexp"
	"strings"
	"strconv"
	"database/sql"
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

func sqlQuery(conn *sql.DB, s sqlstring) (*sql.Rows,error) {
	stmtstr := string(s)
	log.Println("[SQL]", stmtstr)
	stmt, err := conn.Query(stmtstr)
	return stmt, err
}

func sqlQueryRow(conn *sql.DB, s sqlstring) *sql.Row {
	return conn.QueryRow(sql2string(s))
}


// functions to prepare sql statements

func sqlStar(t string) sqlstring{
	return string2sql("SELECT * FROM `" + sqlProtectIdentifier(t) + "`")
}

func sqlSelect(c string, t string) sqlstring{
	return string2sql("SELECT `" + sqlProtectIdentifier(c) + "` FROM `" + sqlProtectIdentifier(t) + "`")
}

func sqlOrder(order string, desc string) sqlstring{
	var query string
	if order != "" {
		query = " ORDER BY `" + sqlProtectIdentifier(order) + "`"
		if desc != "" {
			query = query + " DESC"
		}
	}
	return string2sql(query)
}

// records start with number 1. Every child knows

func sqlLimit(limit int, offset int)sqlstring{
	query := " LIMIT " + strconv.Itoa(maxI(limit,1))
	if offset > 0 {
		query = query + " OFFSET " + strconv.Itoa(offset - 1)
	}
	return string2sql(query)
}


func sqlCount(t string)sqlstring{
	return string2sql("SELECT COUNT(*) FROM `" + sqlProtectIdentifier(t) + "`")
}

func sqlColumns(t string)sqlstring{
	return string2sql("SHOW COLUMNS FROM `" + sqlProtectIdentifier(t) + "`")
}

func sqlInsert(t string)sqlstring{
	return string2sql("INSERT INTO `" + sqlProtectIdentifier(t) + "`")
}

func sqlUpdate(t string)sqlstring{
	return string2sql("UPDATE `" + sqlProtectIdentifier(t) + "`")
}

func sqlDelete(t string)sqlstring{
	return string2sql("DELETE FROM `" + sqlProtectIdentifier(t) + "`")
}

func sqlWhere(k string, c string, v string)sqlstring{
	return string2sql(" WHERE `" + sqlProtectIdentifier(k) + "`" + sqlFilterComparator(c) + "\"" + sqlProtectString(v) + "\"")
}

func sqlWhereClauses(clauses []string)sqlstring{
	return string2sql(" WHERE " + strings.Join(clauses, " && "))
}

func sqlSetClauses(clauses []string)sqlstring{
	return string2sql(" SET " + strings.Join(clauses, " , "))
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

func sqlProtectIdentifier(s string) string {
	if s != "" && strings.ContainsAny(s, "`") {
		r := strings.Replace(s, "`", "``", -1)
		if DEBUGFLAG {
			log.Println("[SQLINJECTION?]", s+" -> "+r)
		}
		return r
	} else {
		return s
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

func sqlProtectString(s string) string {
	if s != "" && strings.ContainsAny(s, "\"") {
		r := strings.Replace(s, "\"", "\"\"", -1)
		if DEBUGFLAG {
			log.Println("[SQLINJECTION?]", s+" -> "+r)
		}
		return r
	} else {
		return s
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

func sqlFilterNumber(t string) string {
	re := regexp.MustCompile("^ *(" + SQLNUM + ") *$")
	return re.FindString(t)
}

func sqlFilterComparator(t string) string {
	re := regexp.MustCompile("^ *(" + SQLCMP + ") *$")
	return re.FindString(t)
}
