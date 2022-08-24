package main

import (
	"fmt"
	"github.com/Bennu-Li/notification-restapi/controllers"
	"github.com/Bennu-Li/notification-restapi/docs"
	"github.com/Bennu-Li/notification-restapi/models"
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
	db, err := models.InitMySQL(os.Getenv("MYSQLSERVER"))
	defer db.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = models.CreateTable(db, "./database/db_messagetemplate_mysql.sql")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = models.CreateTable(db, "./database/db_userbehavior_mysql.sql")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// docHost := flag.String("docHost", "localhost:8080", "doc host")
	docHost := os.Getenv("DOCHOST")
	if docHost != "" {
		docs.SwaggerInfo.Host = docHost + ":8080"
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
