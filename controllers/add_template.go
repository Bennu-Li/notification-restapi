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

func AddTemplate(c *gin.Context, db *sql.DB) (int, error) {
	t := &Template{}
	if c.ShouldBind(t) != nil {
		c.String(400, "faild")
	}
	id, err := t.SaveTemplate(c, db)
	return id, err
}

func (t *Template) SaveTemplate(c *gin.Context, db *sql.DB) (int, error) {
	sqlStr := "INSERT INTO message_template(name, message, user) values (?, ?, ?);"
	userName, ok := c.Get("username")
	user := fmt.Sprintf("%v", userName)
	if !ok {
		return 0, fmt.Errorf("found no username")
	}
	id, err := database.InsertData(db, sqlStr, t.Name, t.Message, user)
	return id, err
}
