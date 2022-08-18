package controllers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

type Template struct {
	// Name          string   `json:"name" form:"name"`
	Message       string `json:"message" form:"message"`
	MessageParams string `json:"messageparam" form:"messageparam"`
}

func AddTemplate(c *gin.Context, db *sql.DB) error {
	t := Template{}
	if c.ShouldBind(&t) != nil {
		c.String(400, "faild")
	}
	err := t.SaveTemplate(db)
	return err
}

func (t *Template) SaveTemplate(db *sql.DB) error {
	return nil
}
