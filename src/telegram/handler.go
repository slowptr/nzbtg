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

	folderName := ""
	log.Println("folderName: " + folderName)
	for {
		status, err := nzb.GetQueue()
		if err != nil {
			msg := tgbotapi.NewEditMessageText(u.Message.Chat.ID, editable.MessageID, "unable to get queue status")
			t.bot.Send(msg)
			continue
		}

		if len(status.Slots) == 0 {
			msg := tgbotapi.NewEditMessageText(u.Message.Chat.ID, editable.MessageID, "100%, download finished...")
			t.bot.Send(msg)
			break
		}

		currentSlot := status.Slots[0]
		folderName = currentSlot.FileName
		msg := tgbotapi.NewEditMessageText(u.Message.Chat.ID, editable.MessageID, status.Slots[0].Percentage+"%")
		t.bot.Send(msg)

		time.Sleep(2 * time.Second)
	}

	// walk through nzb.DLPath and find folderName
	files, err := ioutil.ReadDir(nzb.DLPath)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("looking for folder: " + folderName)

	found := false
	for {
		if found {
			break
		}

		for _, f := range files {
			log.Printf("found: %s, searching for: %s", f.Name(), folderName)
			if f.Name() != folderName {
				continue
			}

			found = true

			msg := tgbotapi.NewEditMessageText(u.Message.Chat.ID, editable.MessageID, "found folder: "+folderName)
			t.bot.Send(msg)

			src := nzb.DLPath + "\\" + f.Name()
			dst := src + ".zip"
			if f.IsDir() {
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

				msg := tgbotapi.NewEditMessageText(u.Message.Chat.ID, editable.MessageID, "zipping and uploading... "+dst)
				t.bot.Send(msg)

				err = cloud.Upload(dst, u.Message.Document.FileName)
				if err != nil {
					msg := tgbotapi.NewEditMessageText(u.Message.Chat.ID, editable.MessageID, err.Error())
					t.bot.Send(msg)
					return
				}

				msg = tgbotapi.NewEditMessageText(u.Message.Chat.ID, editable.MessageID, "upload finished")
				t.bot.Send(msg)
			}
		}

		time.Sleep(5 * time.Second)
	}
}
