package telegram

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/slowptr/nzbtg/sabnzbd"
	"github.com/slowptr/nzbtg/tgcloud"
)

type Telegram struct {
	bot *tgbotapi.BotAPI
}

func New(apiToken string, doDebug bool) (*Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		return nil, err
	}
	bot.Debug = doDebug

	log.Printf("authorized on account: %s", bot.Self.UserName)

	return &Telegram{bot}, nil
}

func (t *Telegram) Run(nzb *sabnzbd.SABNZBD, cloud *tgcloud.TGCloud) {
	t.handleUpdates(func(u tgbotapi.Update) {
		t.messageHandler(u, nzb, cloud)
	})
}

func (t *Telegram) handleUpdates(f func(u tgbotapi.Update)) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)
	for update := range updates {
		f(update)
	}
}
