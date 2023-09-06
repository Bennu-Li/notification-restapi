package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Bennu-Li/notification-restapi/models"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type EmailParams struct {
	Receiver string `json:"receiver" form:"receiver" binding:"required"`
	Subject  string `json:"subject" form:"subject" binding:"required"`
	Message  string `json:"message" form:"message" binding:"required"`
	Format   string `json:"format" form:"format"`
}

type Type string

const (
	TypeText Type = "text"
	TypeHtml Type = "html"
)

// SendNotification godoc
// @Summary     Send message by email
// @Description Send a message to a specify email address
// @Tags        Send
// @Accept      json
// @Produce     json
// @Param       receiver query    string true  "email address"
// @Param       subject  query    string true  "email subject"
// @Param       message  query    string true  "email message"
// @Param       format   query    string false "email content format, text or html, default text"
// @Success     200      {object} map[string]any
// @Router      /email         [post]
// @Security    Bearer
func Email(c *gin.Context, db *sql.DB) {
	e := &EmailParams{}
	err := c.ShouldBind(e)
	if err != nil {
		fmt.Println(err)
		ReturnErrorBody(c, 1, "Your request parameter invalid.", err)
		return
	}

	reader, err := e.generateRequestBody()
	if err != nil {
		fmt.Println(err)
		ReturnErrorBody(c, 1, "faild to generate request body.", err)
		return
	}

	responce, err := Post(os.Getenv("NOTIFICATIONSERVER"), "application/json", reader)

	// Record send message
	status := fmt.Sprintf("%v", responce["Status"])
	errRecord := RecordBehavior(c, db, "email", e.Message, e.Receiver, status)
	if errRecord != nil {
		fmt.Println("record error: ", errRecord)
	}

	if err != nil {
		fmt.Println(err)
		ReturnErrorBody(c, 1, "faild to send message.", err)
		return
	}
	if status != "200" {
		ReturnErrorBody(c, 1, "faild to send message.", fmt.Errorf("%v", responce["Message"]))
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "success",
		})
	}
	return
}

func (e *EmailParams) generateRequestBody() (io.Reader, error) {
	var requestBody map[string]interface{}

	requestBody, err := ReadJson("./alert/to_email.json")
	if err != nil {
		return nil, err
	}
	email := requestBody["receiver"].(map[string]interface{})["spec"].(map[string]interface{})["email"].(map[string]interface{})
	email["to"] = []string{e.Receiver}
	if e.Format != "html" {
		email["tmplType"] = "text"
	} else {
		email["tmplType"] = "html"
	}

	alerts := requestBody["alert"].(map[string]interface{})["alerts"].([]interface{})[0].(map[string]interface{})
	alerts["annotations"].(map[string]interface{})["message"] = e.Message
	alerts["annotations"].(map[string]interface{})["subject"] = e.Subject

	bytesData, _ := json.Marshal(requestBody)
	reader := bytes.NewReader(bytesData)

	return reader, nil
}

func ReadJson(filename string) (map[string]interface{}, error) {
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

func Post(url string, contentType string, jsonFile io.Reader) (map[string]interface{}, error) {
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
	fmt.Println("RSP:", string(body))
	return responce, nil
}

func RecordBehavior(c *gin.Context, db *sql.DB, mess_type, message, receiver, status string) error {
	sqlStr := "INSERT INTO userBehavior(user, application, mess_type, message, receiver, status) values (?, ?, ?, ?, ?, ?);"
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
	err := models.UserBehavior(db, sqlStr, user, app, mess_type, message, receiver, status)
	return err
}
