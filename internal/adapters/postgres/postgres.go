package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

func NewPostgresDB(host, port, user, password, dbname, sslMode string) (*sql.DB, error) {
	if strings.TrimSpace(sslMode) == "" {
		sslMode = "disable"
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslMode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to PostgreSQL")
	return db, nil
}
