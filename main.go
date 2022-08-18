package main

import (
	"github.com/Bennu-Li/notification-restapi/controllers"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	group := router.Group("/api/v1")
	group.POST("/send", controllers.SendMessage)
	group.POST("/add", controllers.AddTemplate)

	router.Run() // 0.0.0.0:8080
}
