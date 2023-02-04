package sabnzbd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/slowptr/nzbtg/utils"
)

type SABNZBD struct {
	baseURL string
}

func New(host string, port string, APIKey string) (*SABNZBD, error) {
	baseURL := fmt.Sprintf("http://%s:%s/api?apikey=%s&output=json", host, port, APIKey)
	return &SABNZBD{baseURL}, nil
}

func (s *SABNZBD) makeRequest(mode string, params map[string]string) (string, error) {
	resp, err := http.Get(s.baseURL + "&mode=" + mode + utils.ParamsToURL(params))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
