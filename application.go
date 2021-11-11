package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hafnisulun/grafana-webhook/controllers"
	"github.com/hafnisulun/grafana-webhook/wassenger"
	_ "github.com/joho/godotenv/autoload"
)

func init() {
	gin.SetMode(os.Getenv("GIN_MODE"))
}

func main() {
	var router = gin.Default()

	webhookController := new(controllers.WebhookController)
	router.POST("webhook/whatsapp", webhookController.Whatsapp)

	wassenger := new(wassenger.Wassenger)
	router.POST("wassenger/webhook", wassenger.Webhook)

	log.Fatal(router.Run())
}
