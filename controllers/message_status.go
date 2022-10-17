package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Bennu-Li/notification-restapi/models"
	"github.com/gin-gonic/gin"
	"net/http"
	// "os"
	"strconv"
	"strings"
	"time"
)

type MessageStatusParams struct {
	Receiver  string `json:"receiver" form:"receiver" binding:"required"`
	Message   string `json:"message" form:"message"`
	MessageId string `json:"message_id" form:"message_id"`
	Interval  int    `json:"interval" form:"interval"`
}

// SendNotification godoc
// @Summary     Check if message has been read
// @Description Check if the latest message has been read
// @Tags        Check
// @Accept      json
// @Produce     json
// @Param       receiver    query    string true   "email address"
// @Param       message     query    string false  "message content"
// @Param       message_id  query    string false  "message id"
// @Param       interval    query    int    false  "Time range for querying history messages, unit hours, default 10 hours "
// @Success     200         {object} map[string]any
// @Router      /messagestatus     [post]
// @Security    Bearer
func MessageStatus(c *gin.Context, db *sql.DB) {
	m := &MessageStatusParams{}

	err := c.ShouldBind(m)
	if err != nil {
		ReturnErrorBody(c, 1, "Your request parameter invalid.", err)
		return
	}

	// Get Token
	token, err := GenTenantAccessToken(appId, appSecret)
	if err != nil {
		ReturnErrorBody(c, 2, "generate feishu access token faild", err)
		return
	}

	// 根据邮箱在数据库中查找 chat_id
	chatId, err := models.GetChatIdByReceiver(db, "select chatid from receiver_info where receiver = ?", m.Receiver)
	if err != nil {
		fmt.Println("search chat id faild: ", err)
	}

	if chatId == "" {
		//如果未找到 chat_id, 使用 bot 给用户发送消息，建立聊天 channel 获取 chat_id
		//给 user 发送消息获取 chat_id
		chatId, err := sendMessagToUser(m.Receiver, m.Message, token, c)
		if err != nil {
			ReturnErrorBody(c, 3, "send message to the receiver by feishu faild, check the parameter receiver", err)
			return
		}

		//根据邮箱获取 user_id
		userId, err := GetUserIdByEmail(m.Receiver, token)
		if err != nil {
			ReturnErrorBody(c, 4, "Get user_id in feishu by user email faild, check the parameter: receiver", err)
			return
		}
		//将receiver信息存入数据库中
		err = RecordReceiverInfo(c, db, userId, m.Receiver, chatId)
		if err != nil {
			ReturnErrorBody(c, 5, "Record the receiver to db faild", err)
			return
		}

		ReturnResBody(c, 0, true)

		return
	}

	//根据 chatId 查找过去一段时间范围内的最新一条历史信息
	messageId, err := getHistoryMessage(chatId, token, m.Interval)
	if err != nil {
		ReturnErrorBody(c, 6, "Get history message faild", err)
		return
	}
	//查看信息已读信息
	ifRead, err := checkMessageStatus(messageId, token)
	if err != nil {
		ReturnErrorBody(c, 7, "Get history message status faild", err)
		return
	}
	ReturnResBody(c, 0, ifRead)
	return
}

func ReturnErrorBody(c *gin.Context, code int, msg string, err error) {
	data := make(map[string]string)
	data["err_log"] = fmt.Sprintf("%v", err)
	c.JSON(http.StatusBadRequest, gin.H{
		"code": code,
		"msg":  msg,
		"data": data,
	})
}

func ReturnResBody(c *gin.Context, code int, ifRead bool) {
	data := make(map[string]bool)
	data["ifRead"] = ifRead
	c.JSON(http.StatusBadRequest, gin.H{
		"code": code,
		"msg":  "success",
		"data": data,
	})
}

func sendMessagToUser(receiver, message, authToken string, c *gin.Context) (string, error) {
	url := "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=email"
	method := "POST"

	if message == "" {
		userName, _ := c.Get("username")
		// user := fmt.Sprintf("%v", userName)
		message = fmt.Sprintf("%v sends you the message using the bot to establish a channel of conversation between you and the bot", userName)
	}

	httpContent := fmt.Sprintf(`{
	"content": "{\"text\":\"%v\"}",
	"msg_type": "text",
	"receive_id": "%v"
	}`, message, receiver)

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

	// messageId, ok = messageData["message_id"].(string)
	chatId, ok := messageData["chat_id"].(string)
	if !ok {
		return "", fmt.Errorf("Get the message_id faild while sending message by bot")
	}

	return chatId, nil
}

func getHistoryMessage(chatId, authToken string, interval int, pageToken ...string) (string, error) {
	var err error
	intervalTime := -10 * time.Hour
	if interval > 0 {
		timeString := strconv.Itoa(interval) + "h"
		intervalTime, err = time.ParseDuration("-" + timeString)
		if err != nil {
			fmt.Println(err)
			return "", err
		}
	} else if interval < 0 {
		fmt.Println("Invalid time interval")
		return "", fmt.Errorf("Invalid time interval")
	}

	cstSh, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return "", err
	}

	now := time.Now().In(cstSh)
	// intervalTime, _ := time.ParseDuration("-" + interval)
	startTime := fmt.Sprintf("%v", now.Add(intervalTime).Unix())
	endTime := fmt.Sprintf("%v", now.Unix())
	// fmt.Println(startTime, endTime)
	var url string
	if len(pageToken) == 0 {
		url = "https://open.feishu.cn/open-apis/im/v1/messages?container_id=" + chatId + "&container_id_type=chat&end_time=" + endTime + "&page_size=50&start_time=" + startTime
	} else {
		url = "https://open.feishu.cn/open-apis/im/v1/messages?container_id=" + chatId + "&container_id_type=chat&page_size=50&page_token=" + pageToken[0] + "&end_time=" + endTime + "&start_time=" + startTime
	}

	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return "", err
	}

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
	// fmt.Println("body: ", jsonData)

	messData, ok := jsonData["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Read history message data faild")
	}

	hasMore, _ := messData["has_more"].(bool)
	if hasMore {
		fmt.Println("Read the next page")
		pageToken, _ := messData["page_token"].(string)
		messageId, err := getHistoryMessage(chatId, authToken, interval, pageToken)
		return messageId, err
	}

	messageItems, ok := messData["items"].([]interface{})
	if !ok {
		return "", fmt.Errorf("Read history message items faild")
	}

	messageLength := len(messageItems)
	if messageLength == 0 {
		return "", fmt.Errorf("There is no new messages, skip")
	}

	// fmt.Printf("Get %v messages in the past %v \n", messageLength, interval)
	latestItem, ok := messageItems[messageLength-1].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Read the latest history message data faild")
	}

	// 判断最新消息的发送者是否是机器人
	sender, ok := latestItem["sender"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Read the latest history message sender faild")
	}
	senderId, ok := sender["id"].(string)
	if !ok {
		return "", fmt.Errorf("Read the latest history message sender ID faild")
	}
	if senderId != appId {
		return "", fmt.Errorf("The latest messages are not sent by bots, skip")
	}

	messageId, ok := latestItem["message_id"].(string)
	if !ok {
		return "", fmt.Errorf("Read the latest history message ID faild")
	}
	return messageId, nil
}

func checkMessageStatus(messageId, authToken string) (bool, error) {
	url := "https://open.feishu.cn/open-apis/im/v1/messages/" + messageId + "/read_users?user_id_type=user_id"
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
		return false, fmt.Errorf("Read the latest history message status faild")
	}

	items, ok := messData["items"].([]interface{})
	if !ok {
		return false, fmt.Errorf("Read the latest history message items status faild")
	}

	if len(items) == 0 {
		// 消息未读
		return false, nil
	}

	return true, nil
}
