package models

import (
	"time"
)

type TwitPost struct {
	Id       string `json:"id,omitempty"`
	Text    string  `json:"text"`
	CreatedAt time.Time  `json:"created_at"`
}
