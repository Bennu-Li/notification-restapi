package main

import (
	"fmt"
	"github.com/Bennu-Li/notification-restapi/controllers"
	"github.com/Bennu-Li/notification-restapi/docs"
	"github.com/Bennu-Li/notification-restapi/models"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"net/http"
	"os"
)

// @title                      Notification API
// @version                    1.0
// @description                This API is used to send notification.
// @host                       localhost:8080
// @BasePath                   /api/v1
// @securityDefinitions.apikey Bearer
// @in                         header
// @name                       Authorization
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
	err = models.CreateTable(db, "./database/db_userinfo_mysql.sql")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = models.CreateTable(db, "./database/db_receiverinfo_mysql.sql")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	docHost := os.Getenv("DOCHOST")
	if docHost != "" {
		docs.SwaggerInfo.Host = docHost
	}

	router := gin.Default()
	group := router.Group("/api/v1")
	group2 := router.Group("/inner/v1")

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "Welcome to use notification api to send your message!",
		})
	})

	group2.POST("/addUser", func(c *gin.Context) {
		controllers.AddUser(c, db)

	})
	group2.POST("/addTemp", controllers.JWTAuthMiddleware(), func(c *gin.Context) {
		controllers.AddTemplate(c, db)

	})
	group2.GET("/list", controllers.JWTAuthMiddleware(), func(c *gin.Context) {
		controllers.ListAllTemplate(c, db)
	})

	group.POST("/auth", func(c *gin.Context) {
		controllers.AuthHandler(c, db)
	})

	group.POST("/refresh", controllers.JWTAuthMiddleware(), controllers.RefreshHandler)

	group.POST("/sms", controllers.JWTAuthMiddleware(), func(c *gin.Context) {
		controllers.SMS(c, db)
	})

	group.POST("/email", controllers.JWTAuthMiddleware(), func(c *gin.Context) {
		controllers.Email(c, db)
	})

	group.POST("/feishu", controllers.JWTAuthMiddleware(), func(c *gin.Context) {
		controllers.Feishu(c, db)
	})

	group.POST("/call", controllers.JWTAuthMiddleware(), func(c *gin.Context) {
		controllers.Call(c, db)
	})

	group.POST("/messagestatus", controllers.JWTAuthMiddleware(), func(c *gin.Context) {
		controllers.MessageStatus(c, db)
	})

	group.GET("/list", controllers.JWTAuthMiddleware(), func(c *gin.Context) {
		controllers.ListTemplate(c, db)
	})

	router.Run() // 0.0.0.0:8080
}
