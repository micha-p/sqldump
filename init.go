package main

import "io/ioutil"
import "text/template" // TODO change to html/template

var templateFormFields = template.New("formfields")
var templateTable = template.New("table")
var templateError = template.New("error")
var loginPage string

// init is a reserved function!
func initTemplate() {
	templateFormFields = template.New("formfields")
	templateTable = template.New("table")
	templateError = template.New("error")

	in, err := ioutil.ReadFile("html/login.html")
	checkY(err)
	loginPage = string(in)

	in, err = ioutil.ReadFile("html/table.html")
	checkY(err)
	_, err = templateTable.Parse(string(in))
	checkY(err)

	in, err = ioutil.ReadFile("html/error.html")
	checkY(err)
	_, err = templateError.Parse(string(in))
	checkY(err)

	in, err = ioutil.ReadFile("html/formFields.html")
	checkY(err)
	_, err = templateFormFields.Parse(string(in))
	checkY(err)
}
