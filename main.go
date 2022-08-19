package main

import (
	"fmt"
	"github.com/Bennu-Li/notification-restapi/controllers"
	"github.com/Bennu-Li/notification-restapi/database"
	_ "github.com/Bennu-Li/notification-restapi/docs"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"os"
)

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1
// @securitydefinitions.apikey Authentication
// @in header
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

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	group.POST("/auth", controllers.AuthHandler)

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
