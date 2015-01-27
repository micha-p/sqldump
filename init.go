package main

import "io/ioutil"
import "text/template"

var templateHead = template.New("th")
var templateCell = template.New("tc")
var loginPage string

func init(){
    in, err := ioutil.ReadFile("html/login.html")
	loginPage = string(in)
	checkY(err)
	_ , err = templateHead.Parse(templH)
	checkY(err)
	_ , err = templateCell.Parse(templC)
	checkY(err)
}
