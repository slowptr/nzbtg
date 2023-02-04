package main

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	token := os.Getenv("TG_TOKEN")
	if token == "" {
		log.Fatal("TG_TOKEN is not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		document := update.Message.Document
		if document != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("file received: %s", document.FileName))
			bot.Send(msg)
		}
	}
}
