package db

import (
	"context"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"

	"github.com/go-pg/pg/v9"
)

type dbLogger struct {
	Logger echo.Logger
}

func (d dbLogger) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	return c, nil
}

func (d dbLogger) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	d.Logger.Info(q.FormattedQuery())
	return nil
}

var db *pg.DB
var once sync.Once

// Init get db instance with .env parameter . you have to create locally .env file first
// Case gcloud deploy put it in app.yml deploy setting file
// after finish using
func Init(logger echo.Logger) *pg.DB {
	once.Do(func() {
		err := godotenv.Load(".env")
		if err != nil {
			logger.Error("Error loading .env file (godotenv)")
		}

		host := os.Getenv("POSTGRES_HOST")
		port := os.Getenv("POSTGRES_PORT")
		user := os.Getenv("POSTGRES_USER")
		password := os.Getenv("POSTGRES_PASSWORD")
		dbName := os.Getenv("POSTGRES_DB")
		networkType := "tcp"
		addr := host + ":" + port

		if os.Getenv("ENV") != "dev" {
			networkType = "unix"
			addr = "/cloudsql/" + host + "/.s.PGSQL." + port
		}

		db = pg.Connect(&pg.Options{
			Network:  networkType,
			Addr:     addr,
			User:     user,
			Password: password,
			Database: dbName,
		})

		db.AddQueryHook(dbLogger{logger})
	})

	return db
}
