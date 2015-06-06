package main

import (
	"html"
	"log"
	"os"
	"regexp"
	"strings"
)

/* 	Any identifier with double quotes in its name has to be escaped.
 */

func escape(name string, link string) Entry {
	return Entry{Text: html.EscapeString(name), Link: strings.Replace(html.EscapeString(link), "%3d", "=", -1)}
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
		log.Println("[SQLINJECTION?]", s+" -> "+r)
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
		log.Println("[SQLINJECTION?]", s+" -> "+r)
		return r
	} else {
		return s
	}
}

// TODO improve with real regexp
func sqlProtectNumericComparison(t string) string {
	reNumeric := regexp.MustCompile("[^-><=!0-9. eE]*")
	return reNumeric.ReplaceAllString(t, "")
}

func troubleF(filename string) error {
	_, err := os.Stat(filename)
	return err
}

// simple error checker
func checkY(err error) {
	if err != nil {
		log.Println("[ERROR]", err)
		os.Exit(1)
	}
}

// Compose dataSourceName from components and globals
func dsn(user string, pw string, host string, port string, db string) string {
	return user + ":" + pw + "@tcp(" + host + ":" + port + ")/" + db
}

func maxI(a int, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

func minI(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
