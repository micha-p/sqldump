package main

import (
	"log"
	"regexp"
	"strings"
)

/* sql escaping using prepared statements in Go is restricted, as it does not work for identifiers */

func sqlStar(t string) string{
	return "SELECT * FROM `" + sqlProtectIdentifier(t) + "`"
}

func sqlSelect(c string, t string) string{
	return "SELECT `" + sqlProtectIdentifier(c) + "` FROM `" + sqlProtectIdentifier(t) + "`"
}

func sqlOrder(o string, d string)string{
	var query string
	if o != "" {
		query = " ORDER BY `" + sqlProtectIdentifier(o) + "`"
		if d != "" {
			query = query + " DESC"
		}
	}
	return query
}

func sqlCount(t string)string{
	return "SELECT COUNT(*) FROM `" + sqlProtectIdentifier(t) + "`"
}

func sqlColumns(t string)string{
	return "SHOW COLUMNS FROM `" + sqlProtectIdentifier(t) + "`"
}

func sqlInsert(t string)string{
	return "INSERT INTO `" + sqlProtectIdentifier(t) + "`"
}

func sqlUpdate(t string)string{
	return "UPDATE `" + sqlProtectIdentifier(t) + "`"
}

func sqlDelete(t string)string{
	return "DELETE FROM `" + sqlProtectIdentifier(t) + "`"
}

func sqlWhere(k string, c string, v string)string{
	return " WHERE `" + sqlProtectIdentifier(k) + "`" + sqlFilterComparator(c) + "\"" + sqlProtectString(v) + "\""
}

func sqlWhereClauses(clauses []string)string{
	return " WHERE " + strings.Join(clauses, " && ")
}

func sqlSetClauses(clauses []string)string{
	return " SET " + strings.Join(clauses, " , ")
}



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
