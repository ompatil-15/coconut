package model

import "time"

type Secret struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
