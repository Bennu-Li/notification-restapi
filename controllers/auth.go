package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type UserClaim struct {
	UserName string `json:"username"`
	AppName  string `json:"appname"`
	jwt.StandardClaims
}

type UserInfo struct {
	User    string `json:"user" form:"user"`
	AppName string `json:"app" form:"app"`
}

const TokenExpireDuration = time.Hour * 24

// ApplyToken godoc
// @Summary      Apply a authrization token
// @Description  Apply a authrization token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        user   query     string     true  "Use the zilliz email"
// @Param        app    query     string     true  "application name"
// @Success      200    {object}  map[string]any
// @Router       /auth  [post]
func AuthHandler(c *gin.Context) {
	u := &UserInfo{}
	err := c.ShouldBind(u)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "bind params error",
		})
		return
	}

	if len(u.User) < 10 || u.User[len(u.User)-10:len(u.User)] != "zilliz.com" {
		fmt.Println("username:", u.User, u.User[len(u.User)-10:len(u.User)])
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "User should be your email address with '@zilliz.com'",
		})
		return
	}

	tokenString, err := u.GenToken()
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("%v", err),
		})
		return
	}

	err = u.SendToken("Bearer " + tokenString)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("%v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"token": "Bearer " + tokenString,
		"msg":   "The token has been sent to your email address, and the token is valid for one day",
	})
	return
}

func (u *UserInfo) GenToken() (string, error) {
	c := UserClaim{
		u.User,
		u.AppName,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpireDuration).Unix(),
			Issuer:    "notification",
		},
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
	requestBody, err := readJson("./alert/to_email.json")
	if err != nil {
		return err
	}
	email := requestBody["receiver"].(map[string]interface{})["spec"].(map[string]interface{})["email"].(map[string]interface{})
	email["to"] = []string{u.User}

	alerts := requestBody["alert"].(map[string]interface{})["alerts"].([]interface{})[0].(map[string]interface{})
	alerts["annotations"].(map[string]interface{})["message"] = token
	alerts["annotations"].(map[string]interface{})["subject"] = "Notification api token"

	bytesData, _ := json.Marshal(requestBody)
	reader := bytes.NewReader(bytesData)

	responce, err := post(os.Getenv("NOTIFICATIONSERVER"), "application/json", reader)

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
// @Summary      Refresh Token
// @Description  Refresh Token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]any
// @Router       /refresh  [get]
// @Security Bearer
func RefreshHandler(c *gin.Context) {
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

	u := &UserInfo{
		User:    user,
		AppName: app,
	}
	tokenString, err := u.GenToken()
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  fmt.Sprintf("%v", err),
		})
		return
	}

	newToken := "Bearer " + tokenString

	// err = u.SendToken(newToken)
	// if err != nil {
	// 	fmt.Println(err)
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"code": 400,
	// 		"msg":  "send token faild",
	// 	})
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{
		"code":      200,
		"new token": newToken,
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
