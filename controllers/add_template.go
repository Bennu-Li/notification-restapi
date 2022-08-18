package controllers

import (
	"github.com/gin-gonic/gin"
)

type Template struct {
	// Name          string   `json:"name" form:"name"`
	Message       string `json:"message" form:"message"`
	MessageParams string `json:"messageparam" form:"messageparam"`
}

func AddTemplate(c *gin.Context) {
	t := Template{}
	if c.ShouldBind(&t) != nil {
		c.String(400, "faild")
	}

}

func (t *Template) SaveTemplate() {

}
