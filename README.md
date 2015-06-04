# sqlgopher

A small web-based tool for database administration.

- access via login mask or bookmark
- credentials stored in secure cookies
- fast dumping of database content
- inserting and querying data
- potentially changing database driver (TODO)
- templates for html

### Installation

    export GOPATH=$PWD
    go get github.com/go-sql-driver/mysql
    go get github.com/gorilla/securecookie
    go get -u github.com/micha-p/sqlgopher

### Usage

Run server for web interface

    export GOPATH=$PWD
    go run sqlgopher.go init.go dump.go aux.go auth.go table.go action.go cert.go get.go -d -c="html/table.css"

or using a prebuilt binary

    export GOPATH=$PWD
    go build sqlgopher.go init.go dump.go aux.go auth.go table.go action.go cert.go get.go
    ./sqlgopher -d -c="html/table.css"


Access via browser

   [http://localhost:8080](http://localhost:8080)

or more convenient but not secure

   [http://localhost:8080/?user=galagopher&pass=mypassword&host=localhost&port=3306](http://localhost:8080/     user=galagopher&pass=mypassword&host=localhost&port=3306)

or on command line

    w3m 'http://localhost:8080/?user=galagopher&pass=mypassword&host=localhost&port=3306'
    lynx -accept_all_cookies 'http://localhost:8080/?user=galagopher&pass=mypassword&host=localhost&port=3306'
    curl -s 'http://localhost:8080/?user=galagopher&pass=mypassword&host=localhost&port=3306&db=galadb&t=posts' | html2text 

### Security

- no encrypted connection to mysql server
- use only in trusted environments
- insert and query limited by request length
- some data types cause problems at driver level
- passwords might be supplied or bookmarked via URL

#### SQL-injection via Request parameters

To prevent SQL-injection, all supplied identifiers are backqoted and to prevent escaping, all backquotes are escaped by doubling them. 
Values are doublequoted, supplied double quotes are escaped the same way. Where-clauses are especially difficult to ckeck, as this would require full parsing of SQL-expressions. Therefore they are avoided (TODO).


#### Javascript-Injection via Identifiers and Values

If identifiers for tables or fields contain quotes or doublequotes, control might escape from these strings. Therefore they are protected by html-escaping in templates and manually (TODO).

 
#### Login-attack via credentials

Establishing connections to databases is done by the standard library-functions. Trusting in code is required here.  


# License

MIT License
