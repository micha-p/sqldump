package main

import (
	"fmt"
	"net/http"
)

/*
 * <table>
 * <tr> <th>head 1</th> <th>head 2</th> </tr>
 * <tr> <td>cell 1</td> <td>cell 2</td> </tr>
 * <tr> <td>cell 3</td> <td>cell 4</td> </tr>
 * </table>
 */

// TODO create and parse templates at compile time

const tableA = "<table>\n"
const tableO = "</table>\n"
const lineA = "<tr>"
const lineO = "</tr>\n"
const templH = "<th>{{.}}</th>"
const templC = "<td>{{.}}</td>"

type Context struct {
	Title string
	Records   [][]string
}


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
