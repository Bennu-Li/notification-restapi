package controllers

import (
	"database/sql"
	"fmt"
	"github.com/Bennu-Li/notification-restapi/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Template struct {
	Name    string `json:"name" form:"name"`
	Message string `json:"message" form:"message"`
	// Application string  `json:"application" form:"application"`
}

// AddTemplate godoc
// @Summary      Regist message template
// @Description  Add a message template to db
// @Tags         Template
// @Accept       json
// @Produce      json
// @Param        name           query      string  true   "message template name"
// @Param        message        query      string  true   "message template"
// @Success      200  {object}  map[string]any
// @Router       /add  [post]
// @Security Bearer
func AddTemplate(c *gin.Context, db *sql.DB) {
	t := &Template{}
	if c.Bind(t) != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "bind params error",
		})
		return
	}

	if t.Name == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "Template name cannot be empty",
		})
		return
	}

	if t.Message == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "Template message cannot be empty",
		})
		return
	}

	// if t.Application == "" {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"code": 400,
	// 		"msg":  "Application cannot be empty",
	// 	})
	// 	return
	// }

	id, err := t.SaveTemplate(c, db)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("%v", err),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":      200,
			"messageID": id,
			"msg":       "success",
		})
	}

	return
}

func (t *Template) SaveTemplate(c *gin.Context, db *sql.DB) (int, error) {
	sqlStr := "INSERT INTO message_template(name, message, registrant, application) values (?, ?, ?, ?);"
	userName, ok := c.Get("username")
	user := fmt.Sprintf("%v", userName)
	if !ok {
		return 0, fmt.Errorf("The requested user name is not recognized")
	}
	appName, ok := c.Get("appname")
	app := fmt.Sprintf("%v", appName)
	if !ok {
		return 0, fmt.Errorf("The requested app name is not recognized")
	}
	id, err := models.InsertData(db, sqlStr, t.Name, t.Message, user, app)
	return id, err
}
