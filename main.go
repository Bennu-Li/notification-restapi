package main

import (
	"fmt"
	"github.com/Bennu-Li/notification-restapi/controllers"
	"github.com/Bennu-Li/notification-restapi/database"
	"github.com/gin-gonic/gin"
	"os"
)

func main() {
	db, err := database.InitMySQL(os.Getenv("MYSQLSERVER"))
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

	router.POST("/auth", controllers.AuthHandler)

	group.POST("/send", controllers.JWTAuthMiddleware(), func(c *gin.Context) {
		err := controllers.SendMessage(c, db)
		if err != nil {
			fmt.Println(err)
			c.String(400, "faild: %v", err)
		} else {
			c.String(200, "send message successfully!")
		}
	})

	group.POST("/add", controllers.JWTAuthMiddleware(), func(c *gin.Context) {
		id, err := controllers.AddTemplate(c, db)
		if err != nil {
			fmt.Println(err)
			c.String(400, "faild: %v", err)
		} else {
			c.String(200, "add message template successfully, template id: %v", id)
		}
	})

	group.GET("/list", controllers.JWTAuthMiddleware(), func(c *gin.Context) {
		result, err := controllers.ListTemplate(c, db)
		if err != nil {
			fmt.Println(err)
			c.String(400, "faild: %v", err)
		} else {
			c.String(200, "%v", result)
		}
	})

	router.Run() // 0.0.0.0:8080
}
