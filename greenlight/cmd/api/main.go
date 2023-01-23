package main

import (
	"context"
	"database/sql"
	"flag"
	"greenlight.alexedwards.net/internal/data"
	"greenlight.alexedwards.net/internal/jsonlog"
	"greenlight.alexedwards.net/internal/mailer"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		enabled bool
		rps     float64
		burst   int
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://postgres:Yerb0lat0vna@127.0.0.1:5432/greenlight?sslmode=disable", "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "fb97376be9e075", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "e5d5b7ad8a0046", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@greenlight.alexedwards.net>", "SMTP sender")
	//-smtp-host='smtp.gmail.com' -smtp-username="kadyelyer@gmail.com" -smtp-password="Yerb0lat0vna" -smtp-sender="kadyelyer@gmail.com"
	//-smtp-host=smtp.office365.com -smtp-port=587 -smtp-username="211619@astanait.edu.kz" -smtp-password= -smtp-sender="211619@astanait.edu.kz"
	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()

	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}

//migrate -path ./migrations -database 'postgres://postgres:Yerb0lat0vna@127.0.0.1:5432/greenlight?sslmode=disable' up
// /Library/PostgreSQL/14/scripts/runpsql.sh
// -smtp-host="outlook.office365.com" -smtp-username="211619@astanait.edu.kz" -smtp-password="Yerb0lat0vna" -smtp-sender="211396@astanait.edu.kz"
//go run ./cmd/api -smtp-host="smtp.office365.com" -smtp-username="211396@astanait.edu.kz" -smtp-password="Yerb0lat0vna" -smtp-sender="211387@astanait.edu.kz" -smtp-port=587
//go run ./cmd/api -smtp-host="smtp.office365.com" -smtp-username="211396@astanait.edu.kz" -smtp-password="Yerb0lat0vna" -smtp-sender="ilyas.amantayev@gmail.com" -smtp-port=587
//go run ./cmd/api -smtp-host="smtp.office365.com" -smtp-username="211396@astanait.edu.kz" -smtp-password="" -smtp-sender="tleuzhan.mukatayev@astanait.edu.kz" -smtp-port=587
