package controllers

import (
	"bytes"
	// "context"
	"encoding/json"
	"fmt"
	"strings"
	// "reflect"
	"github.com/gobuffalo/packr"
	"io"
	"io/ioutil"
	"net/http"
	// "os"
	"database/sql"
	"github.com/Bennu-Li/notification-restapi/database"
	"github.com/gin-gonic/gin"
)

// var notificationServer string = os.Getenv("NotificationServer")
var notificationServer string = "http://10.12.6.30:19093/api/v2/notifications"

type NotificationParams struct {
	MessageTypeId int          `json:"messagetype" form:"messagetype"`
	MessageParams string       `json:"messageparam" form:"messageparam"`
	ReceiverType  ReceiverType `json:"receivertype" form:"receivertype"`
	Receiver      string       `json:"receiver" form:"receiver"`
}

type ReceiverType string

const (
	ReceiverTypeFeishu ReceiverType = "feishu"
	ReceiverTypeEmail  ReceiverType = "email"
	ReceiverTypeSms    ReceiverType = "sms"
)

func SendMessage(c *gin.Context, db *sql.DB) error {
	s := NotificationParams{}
	if c.ShouldBind(&s) != nil {
		return fmt.Errorf("bind params error")
	}

	requestBody, err := (&s).generateRequestBody(db)
	if err != nil {
		// c.String(400, "faild")
		return err
	}
	bytesData, _ := json.Marshal(requestBody)
	reader := bytes.NewReader(bytesData)

	err = post(notificationServer, "application/json", reader)
	if err != nil {
		// fmt.Println(err)
		// c.String(400, "faild")
		return err
	}
	return nil
	// c.String(200, "send message successfully!")
}

func (s *NotificationParams) mergeMessage(db *sql.DB) string {
	// get message from db by s.MessageTypeId
	// messages := "You verification code is {}, it will"
	messages, _ := database.SearchData(db, "select message from message_template where id = ?", s.MessageTypeId)

	strings.Replace(messages, "{}", s.MessageParams, 1)

	// for i, param := range s.MessageParams {
	// 	strings.Replace(messages, "{}", param, i+1)
	// }

	return messages
}

func (s *NotificationParams) generateRequestBody(db *sql.DB) (map[string]interface{}, error) {
	var requestBody map[string]interface{}
	box := packr.NewBox("alert")
	byteValue := box.String("./alert/alert.json")

	// jsonFile, err := os.Open("alert.json")
	// if err != nil {
	// 	return nil, err
	// }
	// defer jsonFile.Close()
	// byteValue, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal([]byte(byteValue), &requestBody)

	switch s.ReceiverType {
	case "feishu":
		feishu := requestBody["receiver"].(map[string]interface{})["spec"].(map[string]interface{})["feishu"].(map[string]interface{})
		feishu["chatbot"].(map[string]interface{})["webhook"].(map[string]interface{})["value"] = s.Receiver
	case "email":
		email := requestBody["receiver"].(map[string]interface{})["spec"].(map[string]interface{})["email"].(map[string]interface{})
		email["to"].([]interface{})[0] = s.Receiver
	case "sms":
		sms := requestBody["receiver"].(map[string]interface{})["spec"].(map[string]interface{})["sms"].(map[string]interface{})
		sms["phoneNumbers"].([]interface{})[0] = s.Receiver
	}

	alerts := requestBody["alert"].(map[string]interface{})["alerts"].([]interface{})[0].(map[string]interface{})
	// alerts["status"] = ""
	alerts["annotations"].(map[string]interface{})["message"] = s.mergeMessage(db)

	return requestBody, nil
}

func post(url string, contentType string, jsonFile io.Reader) error {
	client := http.Client{}
	rsp, err := client.Post(url, contentType, jsonFile)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	fmt.Println("RSP:", string(body))
	return nil
}
