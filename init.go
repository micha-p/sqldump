package main

import "io/ioutil"
import "text/template"

var templateFormFields = template.New("formfields")
var templateTable = template.New("table")
var templateTableFields = template.New("tablefields")
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

	in, err = ioutil.ReadFile("html/formFields.html")
	checkY(err)
	_, err = templateFormFields.Parse(string(in))
	checkY(err)
}
