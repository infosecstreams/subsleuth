package utils

import "time"

// EventSubsLists represents the structure of our data.
type EventSubsLists struct {
	Subscription []struct {
		Condition struct {
			BroadcasterUserID string `json:"broadcaster_user_id"`
		} `json:"condition"`
		Cost      int       `json:"cost"`
		CreatedAt time.Time `json:"created_at"`
		ID        string    `json:"id"`
		Status    string    `json:"status"`
		Transport struct {
			Callback string `json:"callback"`
			Method   string `json:"method"`
		} `json:"transport"`
		Type    string `json:"type"`
		Version string `json:"version"`
	} `json:"data"`
	Pagination struct {
		Cursor string `json:"cursor"`
	} `json:"pagination"`
	Total int `json:"total"`
}
