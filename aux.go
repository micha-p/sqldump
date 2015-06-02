package main

import (
	"log"
	"os"
	"regexp"
)

func troubleF(filename string) error {
	_, err := os.Stat(filename)
	return err
}

// simple error checker
func checkY(err error) {
	if err != nil {
		log.Println("[ERROR]",err)
		os.Exit(1)
	}
}

func href(base string, name string) string {
	return "<a href=\"" + "?" + base + "\">" + name + "</a>"
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

// TODO improve with real regexp
func sqlFilterNumeric(t string) string {
	reNumeric := regexp.MustCompile("[^-><=!0-9. eE]*")
	return reNumeric.ReplaceAllString(t, "")
}
	
	

	
