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

	id, err := t.SaveTemplate(c, db)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("%v", err),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code":      0,
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

type User struct {
	Name string `json:"name" form:"name"`
	App  string `json:"app" form:"app"`
}

func AddUser(c *gin.Context, db *sql.DB) {
	u := &User{}
	if c.Bind(u) != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "bind params error",
		})
		return
	}

	if u.Name == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "User name cannot be empty",
		})
		return
	}

	id, err := u.SaveUser(c, db)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("%v", err),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"ID":   id,
			"msg":  "success",
		})
	}

	return
}

func (u *User) SaveUser(c *gin.Context, db *sql.DB) (int, error) {
	sqlStr := "INSERT INTO user_info(user, application) values (?, ?);"
	id, err := models.InsertUser(db, sqlStr, u.Name, u.App)
	return id, err
}
