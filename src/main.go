package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/slowptr/nzbtg/config"
)

const CONFIG_PATH = "config.ini"

func main() {
	cfg, err := config.Load(CONFIG_PATH)
	if err != nil {
		config.Create(CONFIG_PATH)
		log.Fatal("couldn't load config file, now creating...")
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.APIToken)
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
