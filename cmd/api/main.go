package main

import (
	"api.go-rifqio.my.id/internal/data"
	newLogger "api.go-rifqio.my.id/internal/logger"
	"api.go-rifqio.my.id/internal/smtp"
	"context"
	"database/sql"
	"flag"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

const version = "1.0.0"

type config struct {
	port string
	env  string
	db   struct {
		dsn string
	}

	limiter struct {
		rps     float64
		burst   int
		enabled bool
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
	logger *newLogger.Logger
	models data.Models
	mailer smtp.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config
	dbUrl := getEnv("POSTGRES_URL")
	smtpPort := getEnv("SMTP_PORT")

	cfg.smtp.host = getEnv("SMTP_HOST")
	cfg.smtp.port, _ = strconv.Atoi(smtpPort)
	cfg.smtp.username = getEnv("SMTP_USERNAME")
	cfg.smtp.password = getEnv("SMTP_PASSWORD")
	cfg.smtp.sender = getEnv("SMTP_SENDER")

	flag.StringVar(&cfg.port, "smtp_port", "localhost:4000", "API server smtp_port")
	flag.StringVar(&cfg.env, "env", "dev", "App environment (dev|staging|prod)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", dbUrl, "PostgreSQL DSN")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limit per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limit max burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enable", true, "Rate limit enabler")

	flag.Parse()

	// Create a new logger instance
	logger := newLogger.New(os.Stdout, newLogger.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()
	logger.PrintInfo("Database connection established!", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: smtp.New(
			cfg.smtp.host,
			cfg.smtp.port,
			cfg.smtp.username,
			cfg.smtp.password,
			cfg.smtp.sender,
		),
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

	loggerAdapter := zerologadapter.New(zerolog.New(os.Stdout))
	db = sqldblogger.OpenDriver(
		cfg.db.dsn,
		db.Driver(),
		loggerAdapter,
		sqldblogger.WithSQLQueryAsMessage(true),
	)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(2)
	db.SetConnMaxIdleTime(4 * time.Hour)
	// Create a context with 5 second timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// use PingContext() to establish conection, if there's no respond within 5 second
	// it will return error
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func getEnv(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error parsing env file")
	}
	return os.Getenv(key)
}
