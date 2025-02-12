package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/joho/godotenv"
	"github.com/tsawler/bookings/internal/config"
	"github.com/tsawler/bookings/internal/driver"
	"github.com/tsawler/bookings/internal/handlers"
	"github.com/tsawler/bookings/internal/helpers"
	"github.com/tsawler/bookings/internal/models"
	"github.com/tsawler/bookings/internal/render"
)

const portNumber = ":8081"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

// main is the main application function
func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	defer close(app.MailChan)
	
	fmt.Println("Starting Email Listener....")
	listenForMail()

	fmt.Println(fmt.Sprintf("Starting application on port %s", portNumber))

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	// what am I going to put in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	

	// read flags
	inProduction := flag.String("production", os.Getenv("PRODUCTION"), "Application is in production")
	useCache := flag.String("cache", os.Getenv("CACHE"), "Use template cache")
	dbHost := flag.String("dbhost", os.Getenv("DBHOST"), "Database host")
	dbName := flag.String("dbname", os.Getenv("DBNAME"), "Database name")
	dbUser := flag.String("dbuser", os.Getenv("DBUSER"), "Database user")
	dbPass := flag.String("dbpass", os.Getenv("DBPASS"), "Database password")
	dbPort := flag.String("dbport", os.Getenv("DBPORT"), "Database port")
	dbSSL := flag.String("dbssl", os.Getenv("DBSSL"), "Database ssl settings (disable, prefer, require)")


	flag.Parse()

	if *dbName == "" || *dbUser == "" {
		fmt.Println("Missing required flags")
		os.Exit(1)
	}

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	// change this to true when in production
	app.InProduction, _ = strconv.ParseBool(*inProduction)
  app.UseCache, _  = strconv.ParseBool(*useCache)
	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// connect to database
	log.Println("Connecting to database...")
	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", *dbHost, *dbPort, *dbName, *dbUser, *dbPass, *dbSSL)
	db, err := driver.ConnectSQL(connectionString)
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
	}
	log.Println("Connected to database!")

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
		return nil, err
	}

	app.TemplateCache = tc
	

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db, nil
}
