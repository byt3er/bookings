package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/byt3er/bookings/internals/config"
	"github.com/byt3er/bookings/internals/driver"
	"github.com/byt3er/bookings/internals/handlers"
	"github.com/byt3er/bookings/internals/helpers"
	"github.com/byt3er/bookings/internals/models"
	"github.com/byt3er/bookings/internals/render"

	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

// main is the main function
func main() {

	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	// close the mailChan channel
	defer close(app.MailChan)
	listenForMail()

	// try sending an email message
	// msg := models.MailData{}
	// msg.To = "jhon@do.ca"
	// msg.From = "me@here.com"
	// msg.Subject = "Some subject"
	// msg.Content = "<h1>Hello, World!</h1>"
	// app.MailChan <- msg

	fmt.Println(fmt.Sprintf("Staring application on port %s", portNumber))

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func run() (*driver.DB, error) {
	//What am I going to put in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})

	// read flags
	inProduction := flag.Bool("production", true, "Application is in production")
	useCache := flag.Bool("cache", true, "Use template cache")
	dbHost := flag.String("dbhost", "localhost", "Database host")
	dbName := flag.String("dbname", "", "Database name")
	dbUser := flag.String("dbuser", "", "Database user")
	dbPass := flag.String("dbpass", "", "Database password")
	dbPort := flag.String("dbport", "5432", "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database ssl settings (disable, prefer, require")

	flag.Parse()
	if *dbName == "" || *dbUser == "" {
		fmt.Println("Missing requied flag to start the application")
		// stop the application
		os.Exit(1)
	}

	// for the blockMap in the admin/reservation-calendar
	gob.Register(map[string]int{})

	// create a channel
	mailChain := make(chan models.MailData)
	// make is avaiable to every part of the application
	app.MailChan = mailChain

	// change this to true when in production
	app.InProduction = *inProduction

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime) // print on the terminal window
	app.InfoLog = infoLog

	// log.Lshortfile gives details about the error
	errorLog = log.New(os.Stdout, "Error\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog
	// set up the session
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	//====================================================
	app.Session = session
	//====================================================
	// connect to the database
	log.Println("Connecting to database....")
	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		*dbHost,
		*dbPort,
		*dbName,
		*dbUser,
		*dbPass,
		*dbSSL,
	)
	//"host=localhost port=5432 dbname=bookings user=postgres password=manoj"
	db, err := driver.ConnectSQL(connectionString)
	if err != nil {
		log.Fatal("Cannot connect to the database! Dying...")
	}
	log.Println("Connect to database!")

	// ======================================================

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache", err)
		return nil, err
	}

	app.TemplateCache = tc
	app.UseCache = *useCache

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)

	// initialize the renderer
	render.NewRenderer(&app)

	// initialize the helper
	helpers.NewHelpers(&app)

	return db, nil
}
