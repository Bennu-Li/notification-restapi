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
	Message  string `json:"message"  form:"message" binding:"required"`
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
// @Param       message        query    string true  "message content"
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

	userName, ok := c.Get("username")
	if ok {
		s.Message = s.Message + "   -- from " + fmt.Sprintf("%v", userName)
	}

	// Get Token
	token, err := GenTenantAccessToken(appId, appSecret)
	if err != nil {
		ReturnErrorBody(c, 2, "faild to generate feishu access token", err)
		return
	}

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

	err = RecordBehavior(c, db, "feishu", s.Message, s.Receiver, "200")
	if err != nil {
		fmt.Println("Error: faild to record user behavior to db: ", err)
	}

	// userId, err := GetUserIdByEmail(s.Receiver, token)
	// if err != nil {
	// 	ReturnErrorBody(c, 4, "faild to get user_id in feishu by receiver_email, check the parameter: receiver", err)
	// 	return
	// }

	// err = RecordReceiverInfo(c, db, userId, s.Receiver, 0)
	// if err != nil {
	// 	fmt.Println("Error: faild to record receiver info: ", err)
	// }

	return
}

func (s *SendParams) sendMessagToUser(authToken string) error {
	url := "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=email"
	method := "POST"

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

	return nil
}
