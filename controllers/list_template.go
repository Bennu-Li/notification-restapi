package controllers

import (
	"database/sql"
	"fmt"
	"github.com/Bennu-Li/notification-restapi/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

// listMessageTemplate godoc
// @Summary      List message template
// @Description  List all message template
// @Tags         Template
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /list  [get]
// @Security Bearer
func ListTemplate(c *gin.Context, db *sql.DB) {
	sqlStr := "select * from message_template where registrant = ?"
	userName, ok := c.Get("username")
	if ok {
		user := fmt.Sprintf("%v", userName)
		result, err := models.GetUserTemplate(db, sqlStr, user)

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
	}
	return
}
