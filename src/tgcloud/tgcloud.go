package tgcloud

import (
	"log"
	"os/exec"
)

type TGCloud struct {
	user   string
	chatID string
}

func New(user string, chatID string) (*TGCloud, error) {
	return &TGCloud{user, chatID}, nil
}

func (t *TGCloud) Upload(filepath string, caption string) error {
	out, err := exec.Command("tgcloud", "-m", "upload", "-n", t.user, "-p", filepath, "-u", t.chatID, "-c", caption).Output()
	if err != nil {
		log.Fatal(err.Error() + " | " + string(out))
		return err
	}
	return nil
}
