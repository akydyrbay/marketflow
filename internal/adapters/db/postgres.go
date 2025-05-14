package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func ConnectToDB() *sql.DB {
	var db *sql.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", "postgres://postgres:postgres@db:5432/leetdb?sslmode=disable")
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Println("Waiting for database to be ready...Try: ", i)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("cannot connect to database:", err)
	}

	return db
}
