package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Bennu-Li/notification-restapi/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type UserClaim struct {
	UserName string `json:"username"`
	AppName  string `json:"appname"`
	jwt.StandardClaims
}

type UserInfo struct {
	User    string `json:"user" form:"user" `
	AppName string `json:"app" form:"app" `
	Send    bool   `json:"send" form:"send"`
	Expires int    `json:"expiration" form:"expiration"`
}

// const TokenExpireDuration = time.Hour * 24

// ApplyToken godoc
// @Summary     Apply a authrization token
// @Description Apply a authrization token
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       user        query    string  true   "email address"
// @Param       app         query    string  true   "application name"
// @Param       send        query    bool    false  "if send token to email, default false"
// @Param       expiration  query    int     false  "token expiry date, unit hours. Maximum: 72, default 24 hours."
// @Success     200  {object} map[string]any
// @Router      /auth  [post]
func AuthHandler(c *gin.Context, db *sql.DB) {
	u := &UserInfo{}
	err := c.ShouldBind(u)
	if err != nil {
		ReturnErrorBody(c, 1, "Your request parameter invalid.", err)
		return
	}

	if u.User == "" || u.AppName == "" {
		ReturnErrorBody(c, 1, "Your request parameter invalid.", fmt.Errorf("Error: Field validation for user or app failed, it can not be null"))
		return
	}

	ok, err := models.CheckUserAuth(db, "select count(*) from user_info where user = ? and application = ?", u.User, u.AppName)
	if err != nil {
		ReturnErrorBody(c, 1, "Faild to check user authorization", err)
		return
	}
	if !ok {
		ReturnErrorBody(c, 1, "You do not have permission to request a token", fmt.Errorf(""))
		return
	}

	var tokenExpireDuration time.Duration

	if u.Expires != 0 && u.Expires != -1 {
		if u.Expires < 0 || u.Expires > 72 {
			ReturnErrorBody(c, 1, "The expiration parameter should be between 0 and 72", fmt.Errorf(""))
			return
		}
		timeString := strconv.Itoa(u.Expires) + "h"
		tokenExpireDuration, _ = time.ParseDuration(timeString)
	} else {
		tokenExpireDuration, _ = time.ParseDuration("24h")
	}

	tokenString, err := u.GenToken(tokenExpireDuration)
	if err != nil {
		ReturnErrorBody(c, 1, "faild to generate token", err)
		fmt.Println(err)
		return
	}

	if u.Send {
		// fmt.Println("Send token to email")
		err = u.SendToken("Bearer Token: " + tokenString)
		if err != nil {
			fmt.Println("Send token faild: ", err)
		}
	}

	data := make(map[string]string)
	data["token"] = tokenString
	if u.Expires != -1 {
		data["expiration"] = fmt.Sprintf("The token will be expires after %v", tokenExpireDuration)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": data,
	})
	return
}

func (u *UserInfo) GenToken(tokenExpireDuration time.Duration) (string, error) {
	// duration, _ := time.ParseDuration(u.Expires)
	// fmt.Println("duration: ", duration)
	var c UserClaim
	if u.Expires == -1 {
		c = UserClaim{
			u.User,
			u.AppName,
			jwt.StandardClaims{
				Issuer: "notification",
			},
		}
	} else {
		c = UserClaim{
			u.User,
			u.AppName,
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(tokenExpireDuration).Unix(),
				Issuer:    "notification",
			},
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	secret, err := GetSecertKey()
	if err != nil {
		return "", err
	}

	return token.SignedString(secret)
}

func GetSecertKey() ([]byte, error) {
	keyFile, err := os.Open(os.Getenv("KEYFILE"))
	if err != nil {
		return nil, err
	}
	defer keyFile.Close()

	byteValue, err := ioutil.ReadAll(keyFile)
	return byteValue, err

}

func (u *UserInfo) SendToken(token string) error {
	requestBody, err := ReadJson("./alert/to_email.json")
	if err != nil {
		return err
	}
	email := requestBody["receiver"].(map[string]interface{})["spec"].(map[string]interface{})["email"].(map[string]interface{})
	email["to"] = []string{u.User}
	email["tmplType"] = "text"

	alerts := requestBody["alert"].(map[string]interface{})["alerts"].([]interface{})[0].(map[string]interface{})
	alerts["annotations"].(map[string]interface{})["message"] = token
	alerts["annotations"].(map[string]interface{})["subject"] = "Notification api token"

	bytesData, _ := json.Marshal(requestBody)
	reader := bytes.NewReader(bytesData)

	responce, err := Post(os.Getenv("NOTIFICATIONSERVER"), "application/json", reader)

	status := fmt.Sprintf("%v", responce["Status"])
	if err != nil {
		return err
	}
	if status != "200" {
		// fmt.Println(responce)
		return fmt.Errorf(fmt.Sprintf("%v", responce))
	}

	return nil
}

// RefreshToken godoc
// @Summary     Refresh Token
// @Description Refresh Token
// @Tags        Auth
// @Accept      json
// @Produce     json
// @Param       expiration  query    int     false  "token expiry date, unit hours. Maximum: 72, default 24 hours."
// @Param       send        query    bool    false  "if send token to email, default false"
// @Success     200 {object} map[string]any
// @Router      /refresh  [post]
// @Security    Bearer
func RefreshHandler(c *gin.Context) {
	u := &UserInfo{}
	err := c.ShouldBind(u)
	if err != nil {
		ReturnErrorBody(c, 1, "Your request parameter invalid.", err)
		return
	}

	userName, okUser := c.Get("username")
	user := fmt.Sprintf("%v", userName)

	appName, okApp := c.Get("appname")
	app := fmt.Sprintf("%v", appName)
	if !okUser || !okApp {
		fmt.Println("user: ", user, "app: ", app)
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "The requested app name or user name are not recognized",
		})
		return
	}

	u.User = user
	u.AppName = app

	var tokenExpireDuration time.Duration
	if u.Expires != 0 {
		if u.Expires < 0 || u.Expires > 72 {
			c.JSON(http.StatusOK, gin.H{
				"code": 400,
				"msg":  "The expiration parameter should be between 0 and 72",
			})
			return
		}
		tokenExpireDuration, _ = time.ParseDuration(strconv.Itoa(u.Expires) + "h")
	} else {
		tokenExpireDuration, _ = time.ParseDuration("24h")
	}

	tokenString, err := u.GenToken(tokenExpireDuration)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("%v", err),
		})
		return
	}

	if u.Send {
		// fmt.Println("Send token to email")
		err = u.SendToken("Bearer Token: " + tokenString)
		if err != nil {
			fmt.Println("Send token faild: ", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":       200,
		"new token":  tokenString,
		"expiration": fmt.Sprintf("The new token will be expires after %v", tokenExpireDuration),
	})
	return
}

func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusOK, gin.H{
				"code": 2003,
				"msg":  "unauthorized",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusOK, gin.H{
				"code": 2004,
				"msg":  "Error authorized",
			})
			c.Abort()
			return
		}

		mc, err := ParseToken(parts[1])
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusOK, gin.H{
				"code": 2005,
				"msg":  "Invalid token",
			})
			c.Abort()
			return
		}

		c.Set("username", mc.UserName)
		c.Set("appname", mc.AppName)
		c.Next() // use c.Get("username") to gain user info
	}
}

func ParseToken(tokenString string) (*UserClaim, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaim{}, func(token *jwt.Token) (i interface{}, err error) {
		secret, err := GetSecertKey()
		if err != nil {
			return nil, err
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*UserClaim); ok && token.Valid { // 校验token
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
