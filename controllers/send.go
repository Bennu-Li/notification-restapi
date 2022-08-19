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
	MessageTypeId int          `json:"id" form:"id"`
	MessageName   string       `json:"name" form:"name"`
	MessageParams string       `json:"params" form:"params"`
	ReceiverType  ReceiverType `json:"receivertype" form:"receivertype"`
	Receiver      string       `json:"receiver" form:"receiver"`
}

type ReceiverType string

const (
	ReceiverTypeFeishu ReceiverType = "feishu"
	ReceiverTypeEmail  ReceiverType = "email"
	ReceiverTypeSms    ReceiverType = "sms"
)

// SendNotification godoc
// @Summary      Send notification to receiver
// @Description  get string by ID
// @Tags         Notification
// @Accept       json
// @Produce      json
// @Param        MessageTypeId      query      int     false  "id"
// @Param        MessageName        query      string  false  "name"
// @Param        MessageParams      query      string  true  "params"
// @Param        ReceiverType       query      string  true  "receivertype"
// @Param        Receiver           query      string  true  "receiver"
// @Success      200  {object}  map[string]any
// @Router       /send  [post]
// @securitydefinitions.apikey Authentication
// @in header
func SendMessage(c *gin.Context, db *sql.DB) error {
	s := &NotificationParams{}
	if c.ShouldBind(s) != nil {
		return fmt.Errorf("bind params error")
	}

	fmt.Println(s)

	if s.MessageTypeId == 0 && s.MessageName == "" {
		return fmt.Errorf("None of message type id and name")
	}

	requestBody, err := s.generateRequestBody(db)
	if err != nil {
		return err
	}
	fmt.Println(requestBody)
	bytesData, _ := json.Marshal(requestBody)
	reader := bytes.NewReader(bytesData)

	responce, err := post(os.Getenv("NOTIFICARIONSERVER"), "application/json", reader)
	if err != nil {
		return err
	}
	fmt.Println("RSP: {Status:", responce["Status"], ", Message:", responce["Message"], "}")
	status := fmt.Sprintf("%v", responce["Status"])
	if status != "200" {
		return fmt.Errorf(fmt.Sprintf("%v", responce["Message"]))
	}
	return nil
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
	}
	alerts := requestBody["alert"].(map[string]interface{})["alerts"].([]interface{})[0].(map[string]interface{})
	alerts["annotations"].(map[string]interface{})["message"], err = s.mergeMessage(db)
	// alerts["status"] = ""
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
	// get message from db by s.MessageTypeId
	var messages string
	var err error
	if s.MessageTypeId != 0 {
		messages, err = database.SearchData(db, "select message from message_template where id = ?", s.MessageTypeId)
	} else {
		messages, err = database.SearchData(db, "select message from message_template where name = ?", s.MessageName)
	}
	if err != nil {
		return "", err
	}

	// messages = strings.Replace(messages, "{opt}", s.MessageParams, 1)

	params := strings.Split(s.MessageParams, "|")
	count := strings.Count(messages, "{opt}")
	if count != len(params) {
		return "", fmt.Errorf("Error the number of parameters does not match the required by the template message")
	}
	for i, param := range params {
		messages = strings.Replace(messages, "{opt}", param, i+1)
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
