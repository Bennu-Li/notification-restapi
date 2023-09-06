package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
)

type FeishuParams struct {
	Receiver string `json:"receiver" form:"receiver" binding:"required"`
	Message  string `json:"message" form:"message" binding:"required"`
}

// SendNotification godoc
// @Summary     Send message by feishu bot
// @Description Send a message to a feishu group by feishu bot webhook
// @Tags        Send
// @Accept      json
// @Produce     json
// @Param       receiver query    string true "feishu chatbot webhook"
// @Param       message  query    string true "message content"
// @Success     200      {object} map[string]any
// @Router      /group        [post]
// @Security    Bearer
func FeishuGroup(c *gin.Context, db *sql.DB) {
	f := &FeishuParams{}
	err := c.ShouldBind(f)
	if err != nil {
		ReturnErrorBody(c, 1, "Your request parameter invalid.", err)
		return
	}

	userName, ok := c.Get("username")
	if ok {
		f.Message = f.Message + "    -- from " + fmt.Sprintf("%v", userName)
	} else {
		fmt.Println("get userName error")
	}

	reader, err := f.generateRequestBody()
	if err != nil {
		fmt.Println(err)
		ReturnErrorBody(c, 1, "faild to generate request body.", err)
		return
	}

	responce, err := Post(os.Getenv("NOTIFICATIONSERVER"), "application/json", reader)
	// Record send message
	status := fmt.Sprintf("%v", responce["Status"])
	errRecord := RecordBehavior(c, db, "feishuGroup", f.Message, f.Receiver, status)
	if errRecord != nil {
		fmt.Println("record error: ", errRecord)
	}

	if err != nil {
		fmt.Println(err)
		ReturnErrorBody(c, 1, "faild to send message.", err)
		return
	}
	// fmt.Println("RSP: {Status:", responce["Status"], ", Message:", responce["Message"], "}")
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

func (f *FeishuParams) generateRequestBody() (io.Reader, error) {
	var requestBody map[string]interface{}

	requestBody, err := ReadJson("./alert/to_feishu.json")
	if err != nil {
		return nil, err
	}
	feishu := requestBody["receiver"].(map[string]interface{})["spec"].(map[string]interface{})["feishu"].(map[string]interface{})
	feishu["chatbot"].(map[string]interface{})["webhook"].(map[string]interface{})["value"] = f.Receiver

	requestBody["alert"].(map[string]interface{})["alerts"].([]interface{})[0].(map[string]interface{})["annotations"].(map[string]interface{})["message"] = f.Message

	bytesData, _ := json.Marshal(requestBody)
	reader := bytes.NewReader(bytesData)

	return reader, nil
}
