package main

import (
	"fmt"
	"net/http"
)

// TODO create and parse templates at compile time

const tableA = "<table>\n"
const tableO = "</table>\n"
const lineA = "<tr>"
const lineO = "</tr>\n"
const templH = "<th>{{.}}</th>"
const templC = "<td>{{.}}</td>"


func tableHead(w http.ResponseWriter, s string) {
	err := templateHead.Execute(w, s)
	checkY(err)
}

func tableCell(w http.ResponseWriter, s string) {
	err := templateCell.Execute(w, s)
	checkY(err)
}


func tableDuo(w http.ResponseWriter, s1 string, s2 string) {
	fmt.Fprint(w, lineA)
	tableCell(w, s1)
	tableCell(w, s2)
	fmt.Fprint(w, lineO)
}
