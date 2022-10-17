package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Bennu-Li/notification-restapi/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type CallParams struct {
	Receiver  string `json:"receiver_id" form:"receiver_id"`
	Message   string `json:"message" form:"message"`
	MessageId string `json:"message_id" form:"message_id"`
	Retry     int    `json:"retry" form:"retry"`
	Interval  int    `json:"interval" form:"interval"`
}

var (
	appId     = os.Getenv("APP_ID")
	appSecret = os.Getenv("APP_SECRET")
)

// SendNotification godoc
// @Summary     Send an expedited call
// @Description Send an expedited call by feishu
// @Tags        Send
// @Accept      json
// @Produce     json
// @Param       receiver_id    query    string true   "email address"
// @Param       message        query    string false  "message content"
// @Param       message_id     query    string false  "message id"
// @Param       retry          query    int    false  "times of call, unit minutes, default 10 minutes"
// @Param       interval       query    int    false  "repeat call interval"
// @Success     200      {object} map[string]any
// @Router      /call         [post]
// @Security    Bearer
func Call(c *gin.Context, db *sql.DB) {
	var err error
	// Receive params
	call := &CallParams{}
	if c.ShouldBind(call) != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "bind params error",
		})
		return
	}
	// fmt.Println(*call)

	if call.Receiver == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "The parameters receiver cannot be empty",
		})
		return
	}

	// Get Token
	token, err := GenTenantAccessToken(appId, appSecret)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("%v", err),
		})
		return
	}
	// fmt.Println(token)

	// Send a Message to User by Bot to get a messageID
	var chatId string
	if call.MessageId == "" {
		chatId, err = call.sendMessagToUser(token)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": 400,
				"msg":  fmt.Sprintf("%v", err),
			})
			return
		}
	}

	//根据邮箱获取 user_id
	userId, err := GetUserIdByEmail(call.Receiver, token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("%v", err),
		})
		return
	}

	// 发送加急消息
	err = call.callPhone(userId, token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("%v", err),
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "Successfully sent an expedited message by feishu bot",
		})
	}

	fmt.Println("Successfully sent an expedited message by feishu bot")

	err = RecordBehavior(c, db, "Message expedited", call.Receiver, "200")
	if err != nil {
		fmt.Println("Error: record user behavior error: ", err)
	}

	err = RecordReceiverInfo(c, db, userId, call.Receiver, chatId)
	if err != nil {
		fmt.Println("Error: record receiver info error: ", err)
	}

	if call.Retry != 0 {
		go call.reCall(token, userId)
	}

	return
}

func GenTenantAccessToken(appId, appSecret string) (string, error) {
	url := "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"
	method := "POST"
	payload := strings.NewReader("{\"app_id\": \"" + appId + "\", \"app_secret\": \"" + appSecret + "\"}")
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		// fmt.Println(err)
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	jsonData := make(map[string]interface{})
	json.NewDecoder(res.Body).Decode(&jsonData)
	if jsonData["code"].(float64) != 0 {
		err, _ := jsonData["msg"].(string)
		return "", fmt.Errorf(err)
	}
	// fmt.Println("body: ", jsonData)

	token, ok := jsonData["tenant_access_token"]
	if !ok {
		return "", fmt.Errorf("Get Bot access token faild")
	}
	return token.(string), nil
}

func (call *CallParams) sendMessagToUser(authToken string) (string, error) {
	url := "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=email"
	method := "POST"

	if call.Message == "" {
		call.Message = "You receive an expedited message"
	}

	httpContent := fmt.Sprintf(`{
	"content": "{\"text\":\"%v\"}",
	"msg_type": "text",
	"receive_id": "%v"
	}`, call.Message, call.Receiver)

	payload := strings.NewReader(httpContent)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+authToken)
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	jsonData := make(map[string]interface{})
	json.NewDecoder(res.Body).Decode(&jsonData)
	// fmt.Println(jsonData)

	if jsonData["code"].(float64) != 0 {
		err, _ := jsonData["msg"].(string)
		return "", fmt.Errorf(err)
	}

	messageData, ok := jsonData["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Get the message_id faild while sending message with bot")
	}

	call.MessageId, ok = messageData["message_id"].(string)
	if !ok {
		return "", fmt.Errorf("Get the message_id faild while sending message by bot")
	}
	chatId, ok := messageData["chat_id"].(string)
	if !ok {
		return "", fmt.Errorf("Get the chat_id faild while sending message by bot")
	}

	return chatId, nil
}

func GetUserIdByEmail(receiver, authToken string) (string, error) {
	url := "https://open.feishu.cn/open-apis/contact/v3/users/batch_get_id?user_id_type=user_id"
	method := "POST"
	payload := strings.NewReader("{\"emails\": [\"" + receiver + "\"]}")
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+authToken)
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	jsonData := make(map[string]interface{})
	json.NewDecoder(res.Body).Decode(&jsonData)

	if jsonData["code"].(float64) != 0 {
		err, _ := jsonData["msg"].(string)
		return "", fmt.Errorf(err)
	}

	userData, ok := jsonData["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Read the user data faild")
	}

	userList, ok := userData["user_list"].([]interface{})
	if !ok {
		return "", fmt.Errorf("Read the user list faild")
	}

	if len(userList) == 0 {
		return "", fmt.Errorf("found no users for this email")
	}

	user, _ := userList[0].(map[string]interface{})
	userId, _ := user["user_id"].(string)
	return userId, nil
}

func (call *CallParams) callPhone(userId string, authToken string) error {
	url := "https://open.feishu.cn/open-apis/im/v1/messages/" + call.MessageId + "/urgent_phone?user_id_type=user_id"
	method := "PATCH"
	payload := strings.NewReader("{\"user_id_list\": [\"" + userId + "\"]}")

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
	if jsonData["code"].(float64) != 0 {
		err, _ := jsonData["msg"].(string)
		return fmt.Errorf(err)
	}

	return nil
}

func (call *CallParams) checkMessageStatus(authToken string) (bool, error) {
	url := "https://open.feishu.cn/open-apis/im/v1/messages/" + call.MessageId + "/read_users?user_id_type=user_id"
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("Authorization", "Bearer "+authToken)
	res, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	jsonData := make(map[string]interface{})
	json.NewDecoder(res.Body).Decode(&jsonData)
	if jsonData["code"].(float64) != 0 {
		err, _ := jsonData["msg"].(string)
		return false, fmt.Errorf(err)
	}
	// fmt.Println("body: ", jsonData)

	messData, ok := jsonData["data"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("Get the latest history message status faild")
	}

	items, ok := messData["items"].([]interface{})
	if !ok {
		return false, fmt.Errorf("Get the latest history message items status faild")
	}

	if len(items) == 0 {
		// 消息未读
		return false, nil
	}

	return true, nil
}

func (call *CallParams) reCall(token, userId string) {
	var err error
	interval := 600 * time.Second
	if call.Interval > 0 {
		timeString := strconv.Itoa(call.Interval) + "m"
		interval, err = time.ParseDuration(timeString)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else if call.Interval < 0 {
		fmt.Println("Invalid time interval")
		return
	}

	for i := 0; i < call.Retry; i++ {
		time.Sleep(interval)

		ifRead, err := call.checkMessageStatus(token)
		if err != nil {
			fmt.Println(err)
			return
		}

		if ifRead {
			fmt.Println("Messages are read, skip")
			return
		}

		err = call.callPhone(userId, token)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Successfully retry call:", i+1)

	}
	return
}

func RecordReceiverInfo(c *gin.Context, db *sql.DB, userId, receiver, chatId string) error {
	sqlStr := "INSERT INTO receiver_info(receiverid, receiver, chatid) values (?, ?, ?);"
	err := models.ReceiverInfo(db, sqlStr, userId, receiver, chatId)
	return err
}
