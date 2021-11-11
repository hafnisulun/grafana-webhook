package wassenger

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type Wassenger struct{}

type schedule struct {
	Enabled bool `json:"enabled"`
}

type retry struct {
	Count       int    `json:"count"`
	LastRetryAt string `json:"lastRetryAt"`
}

type data struct {
	Id             string   `json:"id"`
	Phone          string   `json:"phone"`
	Status         string   `json:"status"`
	DeliveryStatus string   `json:"deliveryStatus"`
	CreatedAt      string   `json:"createdAt"`
	SentAt         string   `json:"sentAt"`
	FailedAt       string   `json:"failedAt"`
	ProcessedAt    string   `json:"processedAt"`
	WebhookStatus  string   `json:"webhookStatus"`
	Message        string   `json:"message"`
	Priority       string   `json:"priority"`
	Schedule       schedule `json:"schedule"`
	Retry          retry    `json:"retry"`
	Device         string   `json:"device"`
}

type requestParams struct {
	Token   string `form:"token"`
	Account string `form:"account"`
}

type requestBody []messageEvent

type messageEvent struct {
	Event     string `json:"event"`
	Date      string `json:"date"`
	Message   string `json:"message"`
	Entity    string `json:"entity"`
	EntityId  string `json:"entityId"`
	EntityUrl string `json:"entityUrl"`
	User      string `json:"user"`
	Data      data   `json:"data"`
}

type slackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type slackAttachment struct {
	Fallback string       `json:"fallback"`
	Pretext  string       `json:"pretext"`
	Color    string       `json:"color"`
	Fields   []slackField `json:"fields"`
}

type slackRequestBody struct {
	Attachments []slackAttachment `json:"attachments"`
}

// POST wassenger/webhook
// Webhook from Wassenger
func (r Wassenger) Webhook(c *gin.Context) {
	var requestParams requestParams

	if err := c.ShouldBindQuery(&requestParams); err != nil {
		log.Println("[Error] Request params not match, err:", err)
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Unprocessable entity",
		})
		return
	}

	byteRequestParams, err := json.Marshal(requestParams)

	if err != nil {
		log.Println("[Error] Marshal request params failed, err:", err)
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Unprocessable entity",
		})
		return
	}

	log.Println("Request params:", string(byteRequestParams))

	if requestParams.Token != os.Getenv("APP_TOKEN") {
		log.Println("[Error] Token not match")
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Forbidden",
		})
		return
	}

	if requestParams.Account == "" {
		log.Println("[Error] Account empty")
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Forbidden",
		})
		return
	}

	var requestBody requestBody

	if err := c.ShouldBindBodyWith(&requestBody, binding.JSON); err != nil {
		log.Println("[Error] Request body not match, err:", err)
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Unprocessable entity",
		})
		return
	}

	byteRequestBody, err := json.Marshal(requestBody)

	if err != nil {
		log.Println("[Error] Marshal request params failed, err:", err)
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Unprocessable entity",
		})
		return
	}

	log.Println("Request body:", string(byteRequestBody))

	for _, messageEvent := range requestBody {
		slackRequestBody := &slackRequestBody{
			Attachments: []slackAttachment{
				{
					Fallback: messageEvent.Message,
					Pretext:  messageEvent.Message,
					Color:    "good",
					Fields: []slackField{
						{
							Title: "Account",
							Value: requestParams.Account,
							Short: true,
						},
						{
							Title: "Phone",
							Value: messageEvent.Data.Phone,
							Short: true,
						},
						{
							Title: "Status",
							Value: messageEvent.Data.Status,
							Short: true,
						},
						{
							Title: "Message",
							Value: messageEvent.Data.Message,
							Short: true,
						},
					},
				},
			},
		}

		payloadBuf := new(bytes.Buffer)
		json.NewEncoder(payloadBuf).Encode(slackRequestBody)

		log.Println("Slack request body: " + payloadBuf.String())

		res, err := http.Post(os.Getenv("SLACK_WEBHOOK_URL"), "application/json", payloadBuf)

		if err != nil {
			log.Println("[Error] Request to Slack failed, err:", err)
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
				"error": "Unprocessable entity",
			})
			return
		}

		defer res.Body.Close()

		resBodyBytes, err := ioutil.ReadAll(res.Body)

		if err != nil {
			log.Println("[Error] Read response from Slack failed, err:", err)
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
				"error": "Unprocessable entity",
			})
			return
		}

		resBodyString := string(resBodyBytes)
		log.Printf("Slack response code: %d\n", res.StatusCode)
		log.Println("Slack response body: " + resBodyString)
	}

	response := gin.H{
		"message": "Notification to Slack sent",
	}

	c.JSON(http.StatusOK, response)
}
