package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"e-curatif/internal/data"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Struct that holds all configuration settings for the app.
// CL flags will pass by here when the app starts.
type config struct {
	// The ony purpose of env is to announce if the app runs as development
	// or production
	env string

	// maxIdleTime is the duration of each idle connexion before beeing
	// shutdown.
	// maxOpenConns is the max number of open connexions.
	// maxIdleConns is the max number of idle connexions.
	// TODO: Check if needed to pass through database/sql...
	db struct {
		dsn          string // DB DSN
		maxIdleTime  string // MaxConnIdleTime pgx equi
		maxOpenConns int    // MaxConns pgx equiv
		maxIdleConns int    // Not supported by pgx
	}
	// DB port (PSQL default: 5432)
	port string
}

type application struct {
	// DB will make connexion to other packages structs to passe data from
	// database.
	DB *pgxpool.Pool

	// Log used only for terminal debug. JSON log will be made later.
	infoLog  *log.Logger
	errorLog *log.Logger

	// Connexion to data structs.
	source *data.Source
	info   *data.Info

	// Passing templateCache with application so it can be used easely.
	templateCache map[string]*template.Template

	csv *data.CSV
}

// App version will be with github
var (
// version = vsc.Version()
)

func main() {
	var cfg config

	// This flags have default values for dev and prod.
	// flag "port" it's for localhost only!
	flag.StringVar(&cfg.port, "port", ":3001", "E-Curatif server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// The default values are the minimum recommended.
	// see: https://pgtune.leopard.in.ua/ for more info
	// db-dsn string flag is empty on purpose.
	// See: See Makefile to check the DB DSN.
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 20, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 20, "PostgreSQL max open connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max open connections")

	// errorLog for more important errors returned.
	// infoLog for everything else.
	infoLog := log.New(os.Stderr, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Open connexion with PSQL before launching the application
	db, err := openDB(cfg)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	// Initialize template cache before starting application.
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	// application struct instance containing connections to other packages.
	app := &application{
		DB:            db,
		infoLog:       infoLog,
		errorLog:      errorLog,
		source:        &data.Source{InfoLog: infoLog, ErrorLog: errorLog},
		info:          &data.Info{InfoLog: infoLog, ErrorLog: errorLog},
		templateCache: templateCache,
		csv:           &data.CSV{DB: db, InfoLog: infoLog, ErrorLog: errorLog},
	}

	// default parameters to the router.
	srv := &http.Server{
		Addr:         cfg.port,
		Handler:      app.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", cfg.port)
	err = srv.ListenAndServe()
	srv.ErrorLog.Fatal(err)
}

// Open connexion with PSQL.
// TODO: Need to pass some configs:
//   - max conn, max idel etc...
func openDB(cfg config) (*pgxpool.Pool, error) {
	ctx := context.Background()

	db, err := pgxpool.New(ctx, cfg.db.dsn)
	if err != nil {
		return nil, err
	}

        fmt.Println(cfg.db.dsn)
	return db, nil
}
