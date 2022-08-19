package controllers

import (
	"database/sql"
	// "fmt"
	"github.com/Bennu-Li/notification-restapi/database"
	"github.com/gin-gonic/gin"
)

func ListTemplate(c *gin.Context, db *sql.DB) (map[int]database.MessageTemplate, error) {
	sqlStr := "select * from message_template"
	result, err := database.GetAllTemplate(db, sqlStr)
	return result, err
}
