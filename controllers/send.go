package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	// "github.com/Bennu-Li/notification-restapi/models"
	"github.com/gin-gonic/gin"
	"net/http"
	// "os"
	// "strconv"
	"strings"
	// "time"
)

type SendParams struct {
	Receiver string `json:"receiver" form:"receiver" binding:"required"`
	Message  string `json:"message"  form:"message"`
}

// var (
// 	appId     = os.Getenv("APP_ID")
// 	appSecret = os.Getenv("APP_SECRET")
// )

// SendNotification godoc
// @Summary     Send a message to feishu receiver
// @Description Send a message to a feishu receiver by feishu bot
// @Tags        Send
// @Accept      json
// @Produce     json
// @Param       receiver       query    string true   "email address"
// @Param       message        query    string false  "message content"
// @Success     200      {object} map[string]any
// @Router      /feishu         [post]
// @Security    Bearer
func SendMessage(c *gin.Context, db *sql.DB) {
	var err error
	// Receive params
	s := &SendParams{}
	if c.ShouldBind(s) != nil {
		ReturnErrorBody(c, 1, "Your request parameter invalid.", err)
		return
	}

	// Get Token
	token, err := GenTenantAccessToken(appId, appSecret)
	if err != nil {
		ReturnErrorBody(c, 2, "faild to generate feishu access token", err)
		return
	}
	// fmt.Println(token)

	// Send a Message to User by Bot to get a messageID
	// var chatId string

	err = s.sendMessagToUser(token)
	if err != nil {
		ReturnErrorBody(c, 3, "faild to  send message to feishu receiver", err)
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "Success",
		})
	}

	fmt.Println("Successfully sent an message")

	// err = RecordBehavior(c, db, "Send message to feishu receiver", s.Receiver, "200")
	// if err != nil {
	// 	fmt.Println("Error: faild to record user behavior to db: ", err)
	// }

	// err = RecordReceiverInfo(c, db, userId, s.Receiver, 0)
	// if err != nil {
	// 	fmt.Println("Error: faild to record receiver info: ", err)
	// }

	return
}

// func GenTenantAccessToken(appId, appSecret string) (string, error) {
// 	url := "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"
// 	method := "POST"
// 	payload := strings.NewReader("{\"app_id\": \"" + appId + "\", \"app_secret\": \"" + appSecret + "\"}")
// 	client := &http.Client{}
// 	req, err := http.NewRequest(method, url, payload)
// 	if err != nil {
// 		// fmt.Println(err)
// 		return "", err
// 	}
// 	req.Header.Add("Content-Type", "application/json")
// 	res, err := client.Do(req)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer res.Body.Close()

// 	jsonData := make(map[string]interface{})
// 	json.NewDecoder(res.Body).Decode(&jsonData)
// 	if jsonData["code"].(float64) != 0 {
// 		err, _ := jsonData["msg"].(string)
// 		return "", fmt.Errorf(err)
// 	}
// 	// fmt.Println("body: ", jsonData)

// 	token, ok := jsonData["tenant_access_token"]
// 	if !ok {
// 		return "", fmt.Errorf("Get Bot access token faild")
// 	}
// 	return token.(string), nil
// }

func (s *SendParams) sendMessagToUser(authToken string) error {
	url := "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=email"
	method := "POST"

	// if s.Message == "" {
	// 	s.Message = "You receive an expedited message"
	// }

	httpContent := fmt.Sprintf(`{
	"content": "{\"text\":\"%v\"}",
	"msg_type": "text",
	"receive_id": "%v"
	}`, s.Message, s.Receiver)

	payload := strings.NewReader(httpContent)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+authToken)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	jsonData := make(map[string]interface{})
	json.NewDecoder(res.Body).Decode(&jsonData)
	// fmt.Println(jsonData)

	if jsonData["code"].(float64) != 0 {
		err, _ := jsonData["msg"].(string)
		return fmt.Errorf(err)
	}

	// messageData, ok := jsonData["data"].(map[string]interface{})
	// if !ok {
	// 	return "", fmt.Errorf("Get the message_data faild while sending message with bot")
	// }

	// messageId, ok := messageData["message_id"].(string)
	// if !ok {
	// 	return "", fmt.Errorf("Get the message_id faild while sending message by bot")
	// }

	// chatId, ok := messageData["chat_id"].(string)
	// if !ok {
	// 	return "", fmt.Errorf("Get the chat_id faild while sending message by bot")
	// }

	return nil
}

// func RecordReceiverInfo(c *gin.Context, db *sql.DB, userId, receiver, chatId string) error {
// 	sqlStr := "INSERT INTO receiver_info(receiverid, receiver, chatid) values (?, ?, ?);"
// 	err := models.ReceiverInfo(db, sqlStr, userId, receiver, chatId)
// 	return err
// }
