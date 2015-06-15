package main

import "io/ioutil"
import "html/template" // CLAUSE for query=parameter needed
import "path/filepath"

var templateForm = template.New("formfields")
var templateTable = template.New("table")
var templateError = template.New("error")
var loginPage string

// init is a reserved function!
func initTemplate() {
	templateForm  = template.New("formfields")
	templateTable = template.New("table")
	templateError = template.New("error")

	f, _ := filepath.Abs("html/login.html")
	in, err := ioutil.ReadFile(f)
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

	in, err = ioutil.ReadFile("html/form.html")
	checkY(err)
	_, err = templateForm.Parse(string(in))
	checkY(err)
}
