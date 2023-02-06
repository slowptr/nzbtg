package telegram

import (
	"archive/zip"
	"io"
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
)

func checkFile(u tgbotapi.Update) bool {
	if u.Message.Document == nil {
		return false
	}

	// check if file is .nzb
	if u.Message.Document.MimeType != "application/x-nzb" {
		return false
	}

	return true
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
	for {
		status, err := nzb.GetQueue()
		if err != nil {
			msg := tgbotapi.NewEditMessageText(chatID, editMsgID, "unable to get queue status")
			t.bot.Send(msg)
			continue
		}

		if len(status.Slots) == 0 {
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

func findFolder(parent string, search string) string { // we dont check for directory now
	for {
		files, err := ioutil.ReadDir(parent)
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			if f.Name() != search {
				continue
			}

			return filepath.Join(parent, f.Name())
		}
	}
}

func zipFolder(src string, dst string) error { // add splitting into multiple files
	archive, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		header, _ := zip.FileInfoHeader(info)
		header.Name, _ = filepath.Rel(src, path)
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	})

	return nil
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
	src := findFolder(nzb.DLPath, folderName)
	dst := src + ".zip"

	t.bot.Send(tgbotapi.NewEditMessageText(u.Message.Chat.ID, editable.MessageID, "zipping... "+dst))

	zipFolder(src, dst)

	t.bot.Send(tgbotapi.NewEditMessageText(u.Message.Chat.ID, editable.MessageID, "uploading... "+dst))

	err = cloud.Upload(dst, u.Message.Document.FileName)
	if err != nil {
		msg := tgbotapi.NewEditMessageText(u.Message.Chat.ID, editable.MessageID, err.Error())
		t.bot.Send(msg)
		return
	}

	t.bot.Send(tgbotapi.NewEditMessageText(u.Message.Chat.ID, editable.MessageID, "upload finished"))

	err = os.RemoveAll(src)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Remove(dst) // is still occupied, don't wanna sleep / loop tho
	if err != nil {
		log.Println(err)
	}
}
