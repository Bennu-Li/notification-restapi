package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Bennu-Li/notification-restapi/models"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"strings"
)

type SMSParams struct {
	TemplateId    int    `json:"id" form:"id"`
	TemplateName  string `json:"name" form:"name"`
	MessageParams string `json:"params" form:"params"`
	Receiver      string `json:"receiver" form:"receiver"`
	Message       string
}

// SendNotification godoc
// @Summary     Send message by sms
// @Description Send a message to a phone number
// @Tags        Send
// @Accept      json
// @Produce     json
// @Param       id       query    int    false "Message Template Id"
// @Param       name     query    string false "Message Template Name"
// @Param       params   query    string false "Message Params, separated by '|'"
// @Param       receiver query    string true  "Receiver phone number, area code required"
// @Success     200      {object} map[string]any
// @Router      /sms  [post]
// @Security    Bearer
func SMS(c *gin.Context, db *sql.DB) {
	s := &SMSParams{}
	if c.ShouldBind(s) != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "bind params error",
		})
		return
	}

	if s.TemplateId == 0 && s.TemplateName == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "Requires one of template id or name",
		})
		return
	}

	// 需要增加一个判断用户所选模版是否属于用户自己注册的模版

	if s.Receiver == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "receiver cannot be empty",
		})
		return
	}

	userName, _ := c.Get("username")
	user := fmt.Sprintf("%v", userName)

	reader, err := s.generateRequestBody(db, user)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("%v", err),
		})
		return
	}

	responce, err := Post(os.Getenv("NOTIFICATIONSERVER"), "application/json", reader)
	status := fmt.Sprintf("%v", responce["Status"])
	errRecord := RecordBehavior(c, db, s.Message, s.Receiver, status)
	if errRecord != nil {
		fmt.Println("record error: ", errRecord)
	}
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("%v", err),
		})
		return
	}
	fmt.Println("RSP: {Status:", responce["Status"], ", Message:", responce["Message"], "}")
	if status != "200" {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("%v", responce["Message"]),
		})

		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "send successfully",
		})
	}
	return
}

func (s *SMSParams) generateRequestBody(db *sql.DB, userName string) (io.Reader, error) {
	var requestBody map[string]interface{}

	requestBody, err := ReadJson("./alert/to_sms.json")
	if err != nil {
		return nil, err
	}

	s.Message, err = s.mergeMessage(db, userName)
	if err != nil {
		return nil, err
	}

	sms := requestBody["receiver"].(map[string]interface{})["spec"].(map[string]interface{})["sms"].(map[string]interface{})
	sms["phoneNumbers"] = []string{strings.TrimSpace(s.Receiver)}

	alerts := requestBody["alert"].(map[string]interface{})["alerts"].([]interface{})[0].(map[string]interface{})
	alerts["annotations"].(map[string]interface{})["message"] = s.Message

	bytesData, _ := json.Marshal(requestBody)
	reader := bytes.NewReader(bytesData)

	return reader, nil
}

func (s *SMSParams) mergeMessage(db *sql.DB, userName string) (string, error) {
	// get message from db by s.TemplateId
	var messages string
	var err error
	if s.TemplateId != 0 {
		messages, err = models.SearchData(db, "select message from message_template where id = ? and registrant = ?", s.TemplateId, userName)
	} else {
		messages, err = models.SearchData(db, "select message from message_template where name = ? and registrant = ?", s.TemplateName, userName)
	}
	if err != nil {
		return "", err
	}

	params := strings.Split(s.MessageParams, "|")
	count := strings.Count(messages, "{opt}")
	if count != 0 && count != len(params) {
		return "", fmt.Errorf("The number of parameters does not match the required by the template: have %d, want %d", len(params), count)
	}
	for _, param := range params {
		messages = strings.Replace(messages, "{opt}", param, 1)
	}

	return messages, nil
}
