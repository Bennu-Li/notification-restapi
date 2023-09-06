package controllers

import (
	"database/sql"
	"github.com/PagerDuty/go-pagerduty"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

type PagerdutyParams struct {
	Summary  string `json:"summary" form:"summary" binding:"required"`
	Source   string `json:"source" form:"source" binding:"required"`
	Severity string `json:"severity" form:"severity" binding:"required"`
	Details  string `json:"details" form:"details" binding:"required"`
}

// type Type string

// const (
// 	TypeText Type = "text"
// 	TypeHtml Type = "html"
// )

var (
	pagerdutyAuth       = os.Getenv("PAGERDUTY_AUTH")
	pagerdutyRoutingKey = os.Getenv("PAGERDUTY_ROUTING_KEY")
)

// SendNotification godoc
// @Summary     Use Pagerduty to call
// @Description Use Pagerduty to call a person who
// @Tags        Send
// @Accept      json
// @Produce     json
// @Param       summary   query    string true  "the summary of the alert"
// @Param       source    query    string true  "the source of the alert"
// @Param       severity  query    string true  "the severity of the alert"
// @Param       details   query    string true "the details of the alert"
// @Success     200      {object} map[string]any
// @Router      /pagerduty         [post]
// @Security    Bearer
func Pagerduty(c *gin.Context, db *sql.DB) {
	p := &PagerdutyParams{}
	err := c.ShouldBind(p)
	if err != nil {
		log.Fatal(err)
		ReturnErrorBody(c, 1, "Your request parameter invalid.", err)
		return
	}

	client := pagerduty.NewClient(pagerdutyAuth)
	// list users
	var opts pagerduty.ListUsersOptions

	users, err := client.ListUsers(opts)
	if err != nil {
		log.Fatal(err)
	}
	for _, user := range users.Users {
		log.Printf("User: %s", user.Name)
	}

	// send an alert
	var alertOpts pagerduty.V2Event
	alertOpts.RoutingKey = pagerdutyRoutingKey
	alertOpts.Payload = &pagerduty.V2Payload{
		Summary:  p.Summary,
		Source:   p.Source,
		Severity: p.Severity,
		Details:  p.Details,
	}
	alertOpts.Action = "trigger"

	status := "success"
	_, err = client.ManageEvent(&alertOpts)
	if err != nil {
		log.Fatal(err)
		ReturnErrorBody(c, 1, "faild to call by pagerduty.", err)
		status = "faild"
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "success",
		})
	}

	errRecord := RecordBehavior(c, db, "pagerduty", p.Summary, "pagerduty", status)
	if errRecord != nil {
		log.Fatal(err)
	}

	return
}
