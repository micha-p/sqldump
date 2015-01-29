# sqlgopher

A small web-based tool for database administration. Started during Gopher Gala 2015.
Only basic html to allow for browsers at the command line. 

## prepare

    sudo mysqladmin --defaults-file=/etc/mysql/debian.cnf create galadb
    sudo mysql --defaults-file=/etc/mysql/debian.cnf -e "GRANT ALL PRIVILEGES  ON galadb.*  TO 'galagopher'@'localhost' IDENTIFIED BY 'mypassword'  WITH GRANT OPTION;"
    mysql -p"mypassword" -u galagopher galadb -e 'create table posts (title varchar(64) default null, start date default null);'
    mysql -p"mypassword" -u galagopher galadb -e 'insert into posts values("hello","2015-01-01");'
    mysql -p"mypassword" -u galagopher galadb -e 'insert into posts values("more","2015-01-03");'
    mysql -p"mypassword" -u galagopher galadb -e 'insert into posts values("end","2015-01-23");'
    mysql -p"mypassword" -u galagopher galadb -e 'insert into posts set title="four",start="2015-01-04";'
    mysql -p"mypassword" -u galagopher galadb -e 'insert into posts set title="five",start="2015-01-05";'
    mysql -p"mypassword" -u galagopher galadb -e 'insert into posts set title="six",start="2015-01-06";'
    mysql -p"mypassword" -u galagopher galadb -B -e 'select * from posts;'

## install

    export GOPATH=$PWD
    go get github.com/go-sql-driver/mysql
    go get github.com/gorilla/securecookie
    go get -u github.com/micha-p/sqlgopher

## run

    bin/sqlgopher


## usage

[http://localhost:8080](http://localhost:8080)

or more convenient but not secure

[http://localhost:8080/?user=galagopher&pass=mypassword&host=localhost&port=3306](http://localhost:8080/?user=galagopher&pass=mypassword&host=localhost&port=3306)

or on command line

    w3m 'http://localhost:8080/?user=galagopher&pass=mypassword&host=localhost&port=3306'
    micha@micha-GA-78LMT-USB3:~/bin/sqlgopher$ lynx 'http://localhost:8080/?user=galagopher&pass=mypassword&host=localhost&port=3306'

## caveats

- just basic protection against sql injection via database and table names
- However, any users might destroy just their own databases, logged in before
- insert and query limited by request length

# License

MIT License
