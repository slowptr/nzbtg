package main

import (
	"log"

	"github.com/slowptr/nzbtg/config"
	"github.com/slowptr/nzbtg/sabnzbd"
	"github.com/slowptr/nzbtg/telegram"
)

const CONFIG_PATH = "config.ini"
const DEBUG = true

func main() {
	cfg, err := config.Load(CONFIG_PATH)
	if err != nil {
		config.Create(CONFIG_PATH)
		log.Fatal("couldn't load config file, now creating...")
	}

	nzb, err := sabnzbd.New(cfg.SABNZBD.Host, cfg.SABNZBD.Port, cfg.SABNZBD.APIKey, cfg.SABNZBD.DLPath)
	if err != nil {
		log.Fatal(err)
	}

	bot, err := telegram.New(cfg.Telegram.APIToken, DEBUG)
	if err != nil {
		log.Fatal(err)
	}

	bot.Run(nzb)
}
