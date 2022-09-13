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
	Receiver string `json:"receiver" form:"receiver"`
	Message  string `json:"message" form:"message"`
}

// SendNotification godoc
// @Summary      Send message by feishu
// @Description  Send a message to feishu
// @Tags         Send
// @Accept       json
// @Produce      json
// @Param        receiver       query      string  true    "email address"
// @Param        message        query      string  true    "email message"
// @Success      200            {object}   map[string]any
// @Router       /feishu        [post]
// @Security Bearer
func Feishu(c *gin.Context, db *sql.DB) {
	f := &FeishuParams{}
	if c.ShouldBind(f) != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "bind params error",
		})
		return
	}

	if (f.Receiver == "") || (f.Message == "") {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "The parameters receiver and message cannot be empty",
		})
		return
	}

	reader, err := f.generateRequestBody()
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("%v", err),
		})
		return
	}

	responce, err := Post(os.Getenv("NOTIFICATIONSERVER"), "application/json", reader)

	// Record send message
	status := fmt.Sprintf("%v", responce["Status"])
	errRecord := RecordBehavior(c, db, f.Message, f.Receiver, status)
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
	// fmt.Println("RSP: {Status:", responce["Status"], ", Message:", responce["Message"], "}")
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
