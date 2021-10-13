package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type WebhookController struct{}

type AlertStateType string

type EvalMatch struct {
	Value  float32           `json:"value"`
	Metric string            `json:"metric"`
	Tags   map[string]string `json:"tags"`
}

type WebhookNotifierBody struct {
	Title       string            `json:"title"`
	RuleID      int64             `json:"ruleId"`
	RuleName    string            `json:"ruleName"`
	State       AlertStateType    `json:"state"`
	EvalMatches []*EvalMatch      `json:"evalMatches"`
	OrgID       int64             `json:"orgId"`
	DashboardID int64             `json:"dashboardId"`
	PanelID     int64             `json:"panelId"`
	Tags        map[string]string `json:"tags"`
	RuleURL     string            `json:"ruleUrl,omitempty"`
	ImageURL    string            `json:"imageUrl,omitempty"`
	Message     string            `json:"message,omitempty"`
}

type ReqBody struct {
	Token string   `json:"token"`
	To    []string `json:"to"`
	Param []string `json:"param"`
}

// POST webhook/whatsapp
// Send message via Whatsapp
func (r WebhookController) Whatsapp(c *gin.Context) {
	var webhookNotifierBody WebhookNotifierBody

	if err := c.BindJSON(&webhookNotifierBody); err != nil {
		log.Println("[Error] Request body not match, err:", err)
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Unprocessable entity",
		})
		return
	}

	log.Println("webhookNotifierBody:", webhookNotifierBody)
	log.Println("webhookNotifierBody.EvalMatches:", webhookNotifierBody.EvalMatches)
	log.Println("webhookNotifierBody.Tags:", webhookNotifierBody.Tags)

	message := "-"

	if string(webhookNotifierBody.State) != "ok" {
		message = webhookNotifierBody.Message
	}

	metrics := "-"

	for index, evalMatch := range webhookNotifierBody.EvalMatches {
		log.Println("evalMatch.Metric:", evalMatch.Metric)
		log.Println("evalMatch.Value:", evalMatch.Value)

		if index == 0 {
			metrics = ""
		} else {
			metrics += ", "
		}

		metrics += fmt.Sprintf("%s: %f", evalMatch.Metric, evalMatch.Value)

		log.Println("evalMatch.Tags:")
		for key, tag := range evalMatch.Tags {
			log.Println("key:", key, "tag:", tag)
		}
	}

	param := []string{
		webhookNotifierBody.Title,
		message,
		metrics,
	}

	to := strings.Split(os.Getenv("WHATSAPP_DESTINATIONS"), ",")

	reqBody := &ReqBody{
		Token: os.Getenv("DAMCORP_TOKEN"),
		To:    to,
		Param: param,
	}

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(reqBody)

	log.Println("Request: " + payloadBuf.String())

	url := "https://waba.damcorp.id/whatsapp/sendHsm/pawoon_server_alert_v2"

	res, err := http.Post(url, "application/json", payloadBuf)

	if err != nil {
		log.Println("[Error] Request to damcorp failed, err:", err)
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Unprocessable entity",
		})
		return
	}

	defer res.Body.Close()

	resBodyBytes, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Println("[Error] Read response from damcorp failed, err:", err)
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Unprocessable entity",
		})
		return
	}

	resBodyString := string(resBodyBytes)
	log.Println("Response: " + resBodyString)

	response := gin.H{
		"message": "Whatsapp alert sent",
	}

	c.JSON(http.StatusOK, response)
}
