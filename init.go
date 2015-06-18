package main

import "io/ioutil"
import "html/template" // CLAUSE for query=parameter needed
import "path/filepath"

var templateForm, templateTable, templateError *template.Template
var loginPage string

// init is a reserved function!
func initTemplate() {

	f, _ := filepath.Abs("html/login.html")
	in, err := ioutil.ReadFile(f)
	checkY(err)
	loginPage = string(in)

	in, err = ioutil.ReadFile("html/table.html")
	checkY(err)
	templateTable = template.Must(template.New("table").Parse(string(in)))

	in, err = ioutil.ReadFile("html/error.html")
	checkY(err)
	templateError = template.Must(template.New("error").Parse(string(in)))

	in, err = ioutil.ReadFile("html/form.html")
	checkY(err)
	templateForm = template.Must(template.New("form").Parse(string(in)))
}
