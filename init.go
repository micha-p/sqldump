package main

import "io/ioutil"
import "text/template"

var templateTable = template.New("tt")
var templateTableFields = template.New("tf")
var loginPage string
var helpPage string

func init() {
	in, err := ioutil.ReadFile("html/login.html")
	checkY(err)
	loginPage = string(in)
	in, err = ioutil.ReadFile("html/help.html")
	checkY(err)
	helpPage = string(in)

	in, err = ioutil.ReadFile("html/table.html")
	checkY(err)
	_, err = templateTable.Parse(string(in))
	checkY(err)

	in, err = ioutil.ReadFile("html/tableFields.html")
	checkY(err)
	_, err = templateTableFields.Parse(string(in))
	checkY(err)
}
