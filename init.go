package main

import "io/ioutil"
import "html/template"

var templateHead = template.New("th")
var templateCell = template.New("tc")
var templateTable = template.New("tt")
var loginPage string
var helpPage string
var tablePage string

func init(){
    in, err := ioutil.ReadFile("html/login.html")
	checkY(err)
	loginPage = string(in)
    in, err = ioutil.ReadFile("html/help.html")
	checkY(err)
	helpPage = string(in)
    in, err = ioutil.ReadFile("html/table.html")
	checkY(err)
	tablePage = string(in)
	_ , err = templateTable.Parse(tablePage)
	checkY(err)
}
