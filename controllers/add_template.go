package controllers

import (
	"database/sql"
	"fmt"
	"github.com/Bennu-Li/notification-restapi/database"
	"github.com/gin-gonic/gin"
)

type Template struct {
	Name    string `json:"name" form:"name"`
	Message string `json:"message" form:"message"`
	// MessageParams string `json:"messageparam" form:"messageparam"`
}

func AddTemplate(c *gin.Context, db *sql.DB) error {
	t := &Template{}
	if c.ShouldBind(t) != nil {
		c.String(400, "faild")
	}
	err := t.SaveTemplate(db)
	return err
}

func (t *Template) SaveTemplate(db *sql.DB) error {
	fmt.Println("template: ", t)
	sqlStr := "INSERT INTO message_template(name, message) values (?, ?);"
	err := database.InsertData(db, sqlStr, t.Name, t.Message)
	return err
}
