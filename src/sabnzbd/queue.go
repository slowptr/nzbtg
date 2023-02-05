package sabnzbd

import (
	"encoding/json"
)

type Queue struct {
	Version   string `json:"version"`
	Paused    bool   `json:"paused"`
	Quota     string `json:"quota"`
	HaveQuota bool   `json:"have_quota"`
	LeftQuota string `json:"left_quota"`
	Speed     string `json:"speed"`
	Size      string `json:"size"`
	Status    string `json:"status"`
	TimeLeft  string `json:"timeleft"`
	Slots     []Slot `json:"slots"`
}

type Slot struct {
	Index      int    `json:"index"`
	NzoId      string `json:"nzo_id"`
	Priority   string `json:"priority"`
	FileName   string `json:"filename"`
	Password   string `json:"password"`
	Mb         string `json:"mb"`
	Size       string `json:"size"`
	Percentage string `json:"percentage"`
	Status     string `json:"status"`
	TimeLeft   string `json:"timeleft"`
	AvgAge     string `json:"avg_age"`
}

func (s *SABNZBD) GetQueue() (Queue, error) {
	resp, err := s.makeRequest("queue", nil)
	if err != nil {
		return Queue{}, err
	}

	var queue map[string]Queue
	err = json.Unmarshal([]byte(resp), &queue)
	if err != nil {
		return Queue{}, err
	}

	return queue["queue"], nil
}
