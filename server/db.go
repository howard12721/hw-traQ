package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func openDBFromEnv() (*sql.DB, error) {
	user := os.Getenv("NS_MARIADB_USER")
	password := os.Getenv("NS_MARIADB_PASSWORD")
	host := os.Getenv("NS_MARIADB_HOSTNAME")
	port := os.Getenv("NS_MARIADB_PORT")
	database := os.Getenv("NS_MARIADB_DATABASE")

	if port == "" {
		port = "3306"
	}
	if user == "" || host == "" || database == "" {
		return nil, fmt.Errorf("mariadb environment variables are incomplete")
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4,utf8",
		user,
		password,
		host,
		port,
		database,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}
