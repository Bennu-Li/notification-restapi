package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
)

// init MySQL
func InitMySQL() (*sql.DB, error) {
	dsn := "user:password@tcp(127.0.0.1:3306)/dbname"
	db, err := sql.Open("mysql", dsn)
	// defer db.Close()
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func CreateTable(db *sql.DB) error {
	sqlBytes, err := ioutil.ReadFile("databases/db_messagetemplate_mysql.sql")
	if err != nil {
		return err
	}
	sqlTable := string(sqlBytes)
	fmt.Println(sqlTable)
	return nil
}

// Insert message template
func InsertData(db *sql.DB) {
	//
}

// Get message template by messagetype ID
func SearchData(db *sql.DB) {
	//
}
