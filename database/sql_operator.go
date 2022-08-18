package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"os"
	// "time"
)

type MessageTemplate struct {
	Id          int
	Message     string
	CreatedTime string
}

// init MySQL
func InitMySQL(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// tabelFile = "database/db_messagetemplate_mysql.sql"
func CreateTable(db *sql.DB, tabelFile string) error {
	fmt.Println(os.Getwd())
	sqlBytes, err := ioutil.ReadFile(tabelFile)
	if err != nil {
		return err
	}
	sqlTable := string(sqlBytes)
	fmt.Println(sqlTable)
	_, err = db.Exec(sqlTable)
	if err != nil {
		return err
	}
	return nil
}

// Insert message template
func InsertData(db *sql.DB, sqlStr string, args ...interface{}) error {
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		// log.Fatal(err)
		return err
	}
	res, err := stmt.Exec(args)
	// res, err := db.Exec(sqlStr,args)
	if err != nil {
		// log.Fatal(err)
		return err
	}
	fmt.Println(res)
	return nil
}

// Get message template by messagetype ID
func SearchData(db *sql.DB, sqlStr string, args ...interface{}) (string, error) {
	var message string
	err := db.QueryRow(sqlStr, args).Scan(&message)
	if err != nil {
		fmt.Printf("scan failed, err:%v\n", err)
		return "", err
	}
	return message, nil
}

// List all message template
func GetAllTemplate(db *sql.DB, sqlStr string) error {
	rows, err := db.Query(sqlStr)
	if err != nil {
		fmt.Printf("query failed, err:%v\n", err)
		return err
	}
	defer rows.Close()

	var message MessageTemplate

	for rows.Next() {
		err := rows.Scan(&message.Id, &message.Message, &message.CreatedTime)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			// return nil, err
		}
		fmt.Printf("id:%d\n message:%s\n createTime:%s\n\n", message.Id, message.Message, message.CreatedTime)
	}
	return nil

}
