// This is the means by which we actually
// connect our application to the database
package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

//DB holds the database connection pool
// right now it's going to hold a driver for postgres
// but I might want a driver for different database
// at some point in the future.
// by making a struct type, I can put whatever I want
// in the member here by just adding a new member
// or subsituting the value for SQL to *sql.DB to mariaDB
// connetion pool
type DB struct {
	SQL *sql.DB
}

// intializing empty DB type variable
var dbConn = &DB{}

// NOTE :: these three CONSTANT define the nature of
//          my connection pool

// what is the maximum number of open connections I can have
// never have more than 10 connection to the database
// open at any given time
const maxOpenDbConn = 10

// how many connection can remain in the pool but remain idle
const maxIdleDbConn = 5

// what's the maximum lifetime for a database connection
// 5 minutes
const maxDbLifetime = 5 * time.Minute

// create a connection pool for prostgres

func ConnectSQL(dsn string) (*DB, error) {
	d, err := NewDatabase(dsn)
	if err != nil {
		//The panic built-in
		//function stops normal execution of the current
		//goroutine
		panic(err)
	}
	// set parameter that are available
	// for that connection pool
	// that will stop it from growing out of control
	// and will remove idle database connection and
	// return them to the database when they're not being
	// used and ensure that we have a certain lifetime
	// for all of our database connections
	d.SetMaxOpenConns(maxOpenDbConn)
	d.SetMaxIdleConns(maxIdleDbConn)
	d.SetConnMaxLifetime(maxDbLifetime)

	dbConn.SQL = d

	err = testDB(d)
	if err != nil {
		return nil, err
	}
	return dbConn, nil
}

// test the database
// try to ping the database
func testDB(d *sql.DB) error {
	err := d.Ping()
	if err != nil {
		return err
	}
	return nil
}

// create a new database for the application
func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	// test the connection
	if err = db.Ping(); err != nil {
		return nil, err
	}

	// if everything works properly then
	// return db and nil
	return db, nil
}
