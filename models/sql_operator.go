package models

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	// "os"
	// "time"
)

type MessageTemplate struct {
	Id          int
	Name        string
	Template    string
	Registrant  string
	Application string
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

func CreateTable(db *sql.DB, tabelFile string) error {
	sqlBytes, err := ioutil.ReadFile(tabelFile)
	if err != nil {
		return err
	}
	sqlTable := string(sqlBytes)
	// fmt.Println(sqlTable)
	_, err = db.Exec(sqlTable)
	if err != nil {
		return err
	}
	return nil
}

// Insert message template
func InsertData(db *sql.DB, sqlStr string, arg1 string, arg2 string, arg3 string, arg4 string) (int, error) {
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(arg1, arg2, arg3, arg4)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

// Insert User info
func InsertUser(db *sql.DB, sqlStr string, arg1 string, arg2 string) (int, error) {
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(arg1, arg2)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
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
func GetAllTemplate(db *sql.DB, sqlStr string) ([]MessageTemplate, error) {
	rows, err := db.Query(sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []MessageTemplate

	for rows.Next() {
		var message MessageTemplate
		err := rows.Scan(&message.Id, &message.Name, &message.Template, &message.Registrant, &message.Application, &message.CreatedTime, &message.UpdateTime)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return nil, err
		}
		result = append(result, message)
		// fmt.Printf("id: %d\nname: %s\nmessage: %s\ncreateTime: %s\nupdateTime: %s\n\n", message.Id, message.Name, message.Template, message.CreatedTime, message.UpdateTime)
	}
	return result, nil

}

func GetUserTemplate(db *sql.DB, sqlStr string, arg1 string) ([]MessageTemplate, error) {
	rows, err := db.Query(sqlStr, arg1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []MessageTemplate

	for rows.Next() {
		var message MessageTemplate
		err := rows.Scan(&message.Id, &message.Name, &message.Template, &message.Registrant, &message.Application, &message.CreatedTime, &message.UpdateTime)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return nil, err
		}
		result = append(result, message)
		// fmt.Printf("id: %d\nname: %s\nmessage: %s\ncreateTime: %s\nupdateTime: %s\n\n", message.Id, message.Name, message.Template, message.CreatedTime, message.UpdateTime)
	}
	return result, nil

}

//Get template Name by template id
func GetTemplateNameByID(db *sql.DB, sqlStr string, arg1 int) (string, error) {
	var name string
	err := db.QueryRow(sqlStr, arg1).Scan(&name)
	if err != nil {
		fmt.Printf("scan failed, err:%v\n", err)
		return "", err
	}
	return name, nil
}

//Record user behavior
func UserBehavior(db *sql.DB, sqlStr string, arg1 string, arg2 string, arg3 string, arg4 string, arg5 string) error {
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(arg1, arg2, arg3, arg4, arg5)
	if err != nil {
		return err
	}
	return err
}

//Check user auth
func CheckUserAuth(db *sql.DB, sqlStr string, arg1 string) (bool, error) {
	var count int
	err := db.QueryRow(sqlStr, arg1).Scan(&count)
	if err != nil {
		fmt.Printf("scan failed, err:%v\n", err)
		return false, err
	}
	fmt.Println("get use: ", count)
	if count == 0 {
		return false, nil
	} else {
		return true, nil
	}
}
