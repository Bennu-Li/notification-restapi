package main

import (
	"fmt"
	"github.com/Bennu-Li/notification-restapi/controllers"
	"github.com/Bennu-Li/notification-restapi/database"
	_ "github.com/Bennu-Li/notification-restapi/docs"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	// "net/http"
	"os"
)

// @title           Notificcation API
// @version         1.0
// @description     This API is used to send notification.
// @host      localhost:8080
// @BasePath  /api/v1
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
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
	err = database.CreateTable(db, "./database/db_userbehavior_mysql.sql")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	router := gin.Default()
	group := router.Group("/api/v1")

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	group.POST("/auth", controllers.AuthHandler)

	group.GET("/refresh", controllers.JWTAuthMiddleware(), controllers.RefreshHandler)

	group.POST("/send", controllers.JWTAuthMiddleware(), func(c *gin.Context) {
		controllers.SendMessage(c, db)
	})

	group.POST("/add", controllers.JWTAuthMiddleware(), func(c *gin.Context) {
		controllers.AddTemplate(c, db)

	})

	group.GET("/list", controllers.JWTAuthMiddleware(), func(c *gin.Context) {
		controllers.ListTemplate(c, db)
	})

	router.Run() // 0.0.0.0:8080
}
