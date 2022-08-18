package main

import (
	"fmt"
	"github.com/Bennu-Li/notification-restapi/controllers"
	"github.com/Bennu-Li/notification-restapi/database"
	"github.com/gin-gonic/gin"
	"os"
	// "database/sql"
)

var mysqlServer string = "root:H3Y3i44BfA@tcp(10.101.198.215:3306)/my_database"

func main() {
	db, err := database.InitMySQL(mysqlServer)
	defer db.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = database.CreateTable(db, "./database/db_messagetemplate_mysql.sql")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	router := gin.Default()
	group := router.Group("/api/v1")
	// group.POST("/send", controllers.SendMessage)
	group.POST("/send", func(c *gin.Context) {
		err := controllers.SendMessage(c, db)
		if err != nil {
			c.String(400, "faild")
		} else {
			c.String(200, "send message successfully!")
		}
	})
	// group.POST("/add", controllers.AddTemplate)
	group.POST("/add", func(c *gin.Context) {
		err := controllers.AddTemplate(c, db)
		if err != nil {
			c.String(400, "faild")
		} else {
			c.String(200, "send message successfully!")
		}
	})
	router.Run() // 0.0.0.0:8080
}
