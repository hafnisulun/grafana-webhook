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

type Wassanger struct{}

type param struct {
	Token string `form:"token"`
}

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
func (r Wassanger) Webhook(c *gin.Context) {
	var param param

	if err := c.ShouldBindQuery(&param); err != nil {
		log.Println("[Error] Request params not match, err:", err)
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"error": "Unprocessable entity",
		})
		return
	}

	log.Println("token:", param.Token)

	if param.Token != os.Getenv("APP_TOKEN") {
		log.Println("[Error] Token not match")
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

	log.Println("requestBody:", requestBody)

	for _, messageEvent := range requestBody {
		slackRequestBody := &slackRequestBody{
			Attachments: []slackAttachment{
				{
					Fallback: messageEvent.Message,
					Pretext:  messageEvent.Message,
					Color:    "good",
					Fields: []slackField{
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
					},
				},
			},
		}

		payloadBuf := new(bytes.Buffer)
		json.NewEncoder(payloadBuf).Encode(slackRequestBody)

		log.Println("Request: " + payloadBuf.String())

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
		log.Println("Response: " + resBodyString)
	}

	response := gin.H{
		"message": "Notification to Slack sent",
	}

	c.JSON(http.StatusOK, response)
}
