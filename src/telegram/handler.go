package telegram

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/slowptr/nzbtg/sabnzbd"
	"github.com/slowptr/nzbtg/tgcloud"
	"github.com/slowptr/nzbtg/utils"
)

const MAX_SIZE_MB = 1500

func checkFile(u tgbotapi.Update) bool {
	if u.Message.Document == nil {
		return false
	}

	return strings.HasSuffix(u.Message.Document.FileName, ".nzb")

	/* my telegram app does not set the mimetype??
	// check if file is .nzb
	if u.Message.Document.MimeType != "application/x-nzb" {
		return false
	}
	*/
}

func checkPassword(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	fileBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	str := string(fileBytes)
	return strings.Contains(str, "<meta type=\"password\">")
}

func (t *Telegram) updateDownloadMessage(nzb *sabnzbd.SABNZBD, chatID int64, editMsgID int) string {
	folderName := ""

	failedLoop := 0
	for {
		status, err := nzb.GetQueue()
		if err != nil {
			msg := tgbotapi.NewEditMessageText(chatID, editMsgID, "unable to get queue status")
			t.bot.Send(msg)
			continue
		}

		if len(status.Slots) == 0 {
			if folderName == "" {
				if failedLoop > 3 {
					break
				}

				failedLoop += 1
				time.Sleep(2 * time.Second)
				continue
			}
			msg := tgbotapi.NewEditMessageText(chatID, editMsgID, "100%, download finished...")
			t.bot.Send(msg)
			break
		}

		currentSlot := status.Slots[0]
		folderName = currentSlot.FileName
		msg := tgbotapi.NewEditMessageText(chatID, editMsgID, status.Slots[0].Percentage+"%")
		t.bot.Send(msg)

		time.Sleep(2 * time.Second)
	}

	return folderName
}

func (t *Telegram) messageHandler(u tgbotapi.Update, nzb *sabnzbd.SABNZBD, cloud *tgcloud.TGCloud) {
	if !checkFile(u) {
		msg := tgbotapi.NewMessage(u.Message.Chat.ID, "only .nzb files are supported")
		t.bot.Send(msg)
		return
	}

	// download file
	fileURL, err := t.bot.GetFileDirectURL(u.Message.Document.FileID)
	if err != nil {
		msg := tgbotapi.NewMessage(u.Message.Chat.ID, "unable to download file")
		t.bot.Send(msg)
		return
	}

	if !checkPassword(fileURL) {
		msg := tgbotapi.NewMessage(u.Message.Chat.ID, ".nzb has no password, proceeding...")
		t.bot.Send(msg)
		return
	}

	// add to sabnzbd
	_, err = nzb.AddNZBURL(fileURL)
	if err != nil {
		msg := tgbotapi.NewMessage(u.Message.Chat.ID, "unable to add file to SABNZBD: "+fileURL)
		t.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(u.Message.Chat.ID, "downloading...")
	editable, _ := t.bot.Send(msg)

	folderName := t.updateDownloadMessage(nzb, u.Message.Chat.ID, editable.MessageID)
	if folderName == "" {
		folderName = strings.Split(fileURL, "/")[len(strings.Split(fileURL, "/"))-1]
		folderName = strings.Split(folderName, ".")[0]
	}

	log.Println("looking for folder: " + folderName)
	src := utils.FindFolder(nzb.DLPath, folderName)
	dst := src + ".zip"

	t.bot.Send(tgbotapi.NewEditMessageText(u.Message.Chat.ID, editable.MessageID, "zipping... "+dst))

	err = utils.ZipFolder(src+"\\", dst, MAX_SIZE_MB) // works on windows, does it work on linux?
	if err != nil {
		t.bot.Send(tgbotapi.NewEditMessageText(u.Message.Chat.ID, editable.MessageID, "unable to zip: "+dst))
		log.Fatal(err)
	}

	filepath.Walk(nzb.DLPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(info.Name(), folderName+".zip") {
			t.bot.Send(tgbotapi.NewEditMessageText(u.Message.Chat.ID, editable.MessageID, "uploading... "+path))
			err = cloud.Upload(path, u.Message.Document.FileName)
			if err != nil {
				t.bot.Send(tgbotapi.NewEditMessageText(u.Message.Chat.ID, editable.MessageID, err.Error()))
				log.Fatal(err)
			}

			err = os.Remove(path)
			if err != nil {
				log.Println(err)
			}
		}

		return nil
	})

	t.bot.Send(tgbotapi.NewEditMessageText(u.Message.Chat.ID, editable.MessageID, "upload finished"))

	err = os.RemoveAll(src)
	if err != nil {
		log.Fatal(err)
	}
}
