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
	Name        string
	Message     string
	CreatedTime string
	UpdateTime  string
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
func InsertData(db *sql.DB, sqlStr string, arg1 string, arg2 string, arg3 string) (int, error) {
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		// log.Fatal(err)
		return 0, err
	}
	res, err := stmt.Exec(arg1, arg2, arg3)
	// res, err := db.Exec(sqlStr,args)
	if err != nil {
		// log.Fatal(err)
		return 0, err
	}
	id, err := res.LastInsertId()
	// fmt.Println(res.LastInsertId())
	return int(id), err
}

// Get message template by messagetype ID
func SearchData(db *sql.DB, sqlStr string, arg1 interface{}) (string, error) {
	var message string
	err := db.QueryRow(sqlStr, arg1).Scan(&message)
	if err != nil {
		fmt.Printf("scan failed, err:%v\n", err)
		return "", err
	}
	return message, nil
}

// List all message template
func GetAllTemplate(db *sql.DB, sqlStr string) (map[int]MessageTemplate, error) {
	rows, err := db.Query(sqlStr)
	if err != nil {
		// fmt.Printf("query failed, err:%v\n", err)
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]MessageTemplate)

	for rows.Next() {
		var message MessageTemplate
		err := rows.Scan(&message.Id, &message.Name, &message.Message, &message.CreatedTime, &message.UpdateTime)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return nil, err
		}
		result[message.Id] = message
		fmt.Printf("id: %d\nname: %s\nmessage: %s\ncreateTime: %s\nupdateTime: %s\n\n", message.Id, message.Name, message.Message, message.CreatedTime, message.UpdateTime)
	}
	return result, nil

}
