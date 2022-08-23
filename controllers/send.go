package controllers

import (
	// "context"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	// "reflect"
	// "github.com/gobuffalo/packr"
	"database/sql"
	"github.com/Bennu-Li/notification-restapi/database"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// var notificationServer string = os.Getenv("NotificationServer")
// var notificationServer string = "http://10.12.6.30:19093/api/v2/notifications"

type NotificationParams struct {
	TemplateId    int          `json:"id" form:"id"`
	TemplateName  string       `json:"name" form:"name"`
	MessageParams string       `json:"params" form:"params"`
	ReceiverType  ReceiverType `json:"receivertype" form:"receivertype"`
	Receiver      string       `json:"receiver" form:"receiver"`
	Subject       string       `json:"subject" form:"subject"`
}

type ReceiverType string

const (
	ReceiverTypeFeishu ReceiverType = "feishu"
	ReceiverTypeEmail  ReceiverType = "email"
	ReceiverTypeSms    ReceiverType = "sms"
)

// SendNotification godoc
// @Summary      Send notification
// @Description  Send notification to a specify receiver
// @Tags         Send
// @Accept       json
// @Produce      json
// @Param        id                 query      int     false   "Message Template Id"
// @Param        name               query      string  false   "Message Template Name"
// @Param        params             query      string  false   "Message Params"
// @Param        subject            query      string  false   "email subject"
// @Param        receivertype       query      string  true    "ReceiverType"
// @Param        receiver           query      string  true    "Receiver"
// @Success      200                {object}   map[string]any
// @Router       /send  [post]
// @Security Bearer
func SendMessage(c *gin.Context, db *sql.DB) {
	s := &NotificationParams{}
	if c.ShouldBind(s) != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "bind params error",
		})
		return
	}

	if s.TemplateId == 0 && s.TemplateName == "" {
		// return fmt.Errorf("Requires one of template id or name")
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "Requires one of template id or name",
		})
		return
	}

	if (s.ReceiverType == "") || (s.Receiver == "") {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "receivertype and receiver cannot be empty",
		})
		return
	}

	requestBody, err := s.generateRequestBody(db)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("%v", err),
		})
		return
	}
	fmt.Println(requestBody)
	bytesData, _ := json.Marshal(requestBody)
	reader := bytes.NewReader(bytesData)
	responce, err := post(os.Getenv("NOTIFICARIONSERVER"), "application/json", reader)

	status := fmt.Sprintf("%v", responce["Status"])
	errRecord := s.RecordBehavior(c, db, status)
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

func (s *NotificationParams) generateRequestBody(db *sql.DB) (map[string]interface{}, error) {
	var requestBody map[string]interface{}
	var err error
	switch s.ReceiverType {
	case "feishu":
		requestBody, err = readJson("./alert/to_feishu.json")
		if err != nil {
			return nil, err
		}
		feishu := requestBody["receiver"].(map[string]interface{})["spec"].(map[string]interface{})["feishu"].(map[string]interface{})
		feishu["chatbot"].(map[string]interface{})["webhook"].(map[string]interface{})["value"] = s.Receiver
	case "email":
		requestBody, err = readJson("./alert/to_email.json")
		if err != nil {
			return nil, err
		}
		email := requestBody["receiver"].(map[string]interface{})["spec"].(map[string]interface{})["email"].(map[string]interface{})
		email["to"] = []string{s.Receiver}
	case "sms":
		requestBody, err = readJson("./alert/to_sms.json")
		if err != nil {
			return nil, err
		}
		sms := requestBody["receiver"].(map[string]interface{})["spec"].(map[string]interface{})["sms"].(map[string]interface{})
		sms["phoneNumbers"] = []string{strings.TrimSpace(s.Receiver)}
	default:
		return nil, fmt.Errorf("Error receiver type, should be one of them: feishu, email, sms")
	}
	alerts := requestBody["alert"].(map[string]interface{})["alerts"].([]interface{})[0].(map[string]interface{})
	alerts["annotations"].(map[string]interface{})["message"], err = s.mergeMessage(db)

	if s.ReceiverType == "email" {
		if s.Subject == "" {
			return nil, fmt.Errorf("Need subject for email receiver")
		}
		alerts["annotations"].(map[string]interface{})["subject"] = s.Subject
	}

	if err != nil {
		return nil, err
	}

	return requestBody, nil
}

func readJson(filename string) (map[string]interface{}, error) {
	var requestBody map[string]interface{}

	// box := packr.NewBox("alert")
	// byteValue := box.String("./alert/alert.json")

	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal([]byte(byteValue), &requestBody)
	// fmt.Println(requestBody)

	return requestBody, nil

}

func (s *NotificationParams) mergeMessage(db *sql.DB) (string, error) {
	// get message from db by s.TemplateId
	var messages string
	var err error
	if s.TemplateId != 0 {
		messages, err = database.SearchData(db, "select message from message_template where id = ?", s.TemplateId)
	} else {
		messages, err = database.SearchData(db, "select message from message_template where name = ?", s.TemplateName)
	}
	if err != nil {
		return "", err
	}

	// messages = strings.Replace(messages, "{opt}", s.MessageParams, 1)
	// fmt.Println("sourceParams: ", s.MessageParams)

	params := strings.Split(s.MessageParams, "|")

	// fmt.Println("params: ", params, "length: ", len(params))
	count := strings.Count(messages, "{opt}")
	if count != 0 && count != len(params) {
		return "", fmt.Errorf("The number of parameters does not match the required by the template: have %d, want %d", len(params), count)
	}
	for _, param := range params {
		messages = strings.Replace(messages, "{opt}", param, 1)
	}

	return messages, nil
}

func post(url string, contentType string, jsonFile io.Reader) (map[string]interface{}, error) {
	client := http.Client{}
	rsp, err := client.Post(url, contentType, jsonFile)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	var responce map[string]interface{}
	json.Unmarshal([]byte(body), &responce)
	// fmt.Println(responce["Status"], responce["Message"])
	// fmt.Println("RSP:", string(body))
	return responce, nil
}

func (s *NotificationParams) RecordBehavior(c *gin.Context, db *sql.DB, status string) error {
	sqlStr := "INSERT INTO user_behavior(user, application, template, params, message, status) values (?, ?, ?, ?, ?, ?);"
	userName, ok := c.Get("username")
	user := fmt.Sprintf("%v", userName)
	if !ok {
		return fmt.Errorf("The requested user name is not recognized")
	}
	appName, ok := c.Get("appname")
	app := fmt.Sprintf("%v", appName)
	if !ok {
		return fmt.Errorf("The requested app name is not recognized")
	}
	message, err := s.mergeMessage(db)
	if err != nil {
		return err
	}
	if s.TemplateName == "" {
		name, err := database.GetTemplateNameByID(db, "select name from message_template where id = ?", s.TemplateId)
		if err != nil {
			return err
		}
		s.TemplateName = name
	}

	err = database.UserBehavier(db, sqlStr, user, app, s.TemplateName, s.MessageParams, message, status)
	return err
}
