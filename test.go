
package main

import "fmt"
import "net/http"
import "io/ioutil"
import templateT "text/template"
import templateH "html/template"

var templateTestT = templateT.New("testT")
var templateTestH = templateH.New("testH")

type TestInput struct {
	Text        string
	Tag         string
	Link		string
	Query		string
	PathQuery	string
	URI			string
	URISingle	string
	URIDouble	string
	Double 		string
	Single      string
	AttributeDouble  string
	AttributeSingle  string
}


func testHandler(w http.ResponseWriter, r *http.Request) {

	templateTestT = templateT.New("testT")
	templateTestH = templateH.New("testH")

	in, err := ioutil.ReadFile("html/test.html")
	checkY(err)
	_, err = templateTestT.Parse(string(in))
	checkY(err)
	_, err = templateTestH.Parse(string(in))
	checkY(err)

	c := TestInput{
		"hello world",									// regular text
		"<script>document.write(1+1);</script>", 		// Tag injection
		"http://www.example.com", 						// Link
		"t=hardcoded&n=3", 		
		"test?t=hardcoded&n=3", 					
		"http://localhost:8080/test?t=hardcoded&n=3", 
		"http://localhost:8080/test?m=hello'world&n=3",
		"http://localhost:8080/test?m=hello\"world&n=3",
		"hello\"world",									// Double Quote Escape
		"hello'world",									// Single Quote Escape
		"\"onclick=\"this.value=1+1\" \"continue:",		// Double Quote attribute injection
		"'onclick=\"this.value=1+1\" 'continue:",		// Single Quote attribute injection
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")	
	q := r.URL.Query()
	if q.Get("t")=="hardcoded"{
		testHardcoded(w,c)
	} else if q.Get("t")=="texttemplate" {
		testTextTemplate(w,c)
	} else if q.Get("t")=="htmltemplate" {
		testHtmlTemplate(w,c)
	} else {
		fmt.Fprintf(w,"<html><body>")
		fmt.Fprintf(w,"<a href=\"/test?t=hardcoded\"> hardcoded </a><br>")
		fmt.Fprintf(w,"<a href=\"/test?t=texttemplate\">  text/template </a><br>")
		fmt.Fprintf(w,"<a href=\"/test?t=htmltemplate\">  html/template </a><br>")
		fmt.Fprintf(w, "</body></html>")
	}
}



func testTextTemplate(w http.ResponseWriter, c TestInput) {
	err := templateTestT.Execute(w, c)
	checkY(err)
}

func testHtmlTemplate(w http.ResponseWriter, c TestInput) {
	err := templateTestH.Execute(w, c)
	checkY(err)
}


func testHardcoded(w http.ResponseWriter,c TestInput) {

	fmt.Fprintf(w,"<html><body>")

    fmt.Fprintf(w, "<p> " + c.Text + "</p>\n")

	fmt.Fprintf(w, "<p> " + c.Tag + "</p>\n")

	test := "<input type=\"text\" name=unsafe value=\"" + c.Double + "\">" // Double Quote Escape
	fmt.Fprintf(w, "<p> " + test + "</p>\n")

	test = "<input type=\"text\" name=unsafe value='"   + c.Single + "'>" // Single Quote Escape
	fmt.Fprintf(w, "<p> " + test + "</p>\n")
		
	test = "<input type=\"text\" name=unsafe value=\"" + "hello" + c.AttributeDouble + "world" + "\">" // Double Quote attribute injection
	fmt.Fprintf(w, "<p> " + test + "</p>\n")

	test = "<input type=\"text\" name=unsafe value='"  + "hello" + c.AttributeSingle + "world" + "'>" // Single Quote attribute injection
	fmt.Fprintf(w, "<p> " + test + "</p>\n")

	fmt.Fprintf(w, "</body></html>")
}


