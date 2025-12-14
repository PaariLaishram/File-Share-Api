package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2/log"
)

var DB *sql.DB

func ConnectToDatabase() error {
	cfg := mysql.NewConfig()
	cfg.User = os.Getenv("DBUSER")
	cfg.Passwd = os.Getenv("DBPASS")
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = os.Getenv("DBNAME")

	// Get a database handle.
	var err error
	DB, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Error("Error opening mysql connection: ", err.Error())
		return errors.New("Error opening mysql connection: " + err.Error())
	}
	pingErr := DB.Ping()
	if pingErr != nil {
		log.Error("Error pinging db: ", pingErr.Error())
		return errors.New("Error opening mysql connection: " + pingErr.Error())
	}
	fmt.Println("Connected to database")
	return nil
}

func RunInsert(query string, fields []any) (int, error) {
	result, err := DB.Exec(query, fields...)
	if err != nil {
		log.Error("Error in insert operation: ", err.Error())
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Error("Error in getting last inserted id: ", err.Error())
		return 0, err
	}
	return int(id), nil
}

func RunQuery[T any](query string, fields []any, filter []any, item *T) ([]any, error) {
	rows, err := DB.Query(query, filter...)
	if err != nil {
		log.Error("Error running query: ", err.Error())
		return nil, fmt.Errorf("Error running query: %v", err)
	}
	defer rows.Close()
	var result []any
	for rows.Next() {
		if err := rows.Scan(fields...); err != nil {
			return nil, fmt.Errorf("Error scanning fields: %v", err)
		}
		result = append(result, *item)
	}
	return result, nil
}

func RunUpdate(query string, filters []any) (bool, error) {
	_, err := DB.Exec(query, filters...)
	if err != nil {
		log.Error("Error running update query: ", err.Error())
		return false, err
	}
	return true, nil
}
