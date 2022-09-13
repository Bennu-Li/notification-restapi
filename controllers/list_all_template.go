package controllers

import (
	"database/sql"
	"fmt"
	"github.com/Bennu-Li/notification-restapi/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ListAllTemplate(c *gin.Context, db *sql.DB) {
	sqlStr := "select * from message_template"

	result, err := models.GetAllTemplate(db, sqlStr)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("%v", err),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  result,
		})
	}

	return
}
