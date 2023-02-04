package config

import "gopkg.in/ini.v1"

type configSABNZBD struct {
	Host   string
	Port   string
	APIKey string
	DLPath string
}

type configTelegram struct {
	APIToken string
	ChatID   string
}

type configTGCloud struct {
	User string
}

type Config struct {
	SABNZBD  configSABNZBD
	Telegram configTelegram
	TGCloud  configTGCloud
}

func Load(path string) (Config, error) {
	cfgFile, err := ini.Load(path)
	if err != nil {
		return Config{}, err
	}

	var config Config
	sectionSABNZBD := cfgFile.Section("SABNZBD")
	config.SABNZBD.Host = sectionSABNZBD.Key("HOST").String()
	config.SABNZBD.Port = sectionSABNZBD.Key("PORT").String()
	config.SABNZBD.APIKey = sectionSABNZBD.Key("APIKEY").String()
	config.SABNZBD.DLPath = sectionSABNZBD.Key("DLPATH").String()

	sectionTelegram := cfgFile.Section("TELEGRAM")
	config.Telegram.APIToken = sectionTelegram.Key("APITOKEN").String()
	config.Telegram.ChatID = sectionTelegram.Key("CHATID").String()

	sectionTGCloud := cfgFile.Section("TG_CLOUD")
	config.TGCloud.User = sectionTGCloud.Key("USER").String()

	return config, nil
}

func Create(path string) error {
	cfgFile := ini.Empty()
	sectionSABNZBD := cfgFile.Section("SABNZBD")
	sectionSABNZBD.Key("HOST").SetValue("localhost")
	sectionSABNZBD.Key("PORT").SetValue("8080")
	sectionSABNZBD.Key("APIKEY").SetValue("")
	sectionSABNZBD.Key("DLPATH").SetValue("")

	sectionTelegram := cfgFile.Section("TELEGRAM")
	sectionTelegram.Key("APITOKEN").SetValue("")
	sectionTelegram.Key("CHATID").SetValue("")

	sectionTGCloud := cfgFile.Section("TG_CLOUD")
	sectionTGCloud.Key("USER").SetValue("")

	return cfgFile.SaveTo(path)
}
