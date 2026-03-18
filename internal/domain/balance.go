package domain

import "time"

type Balance struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	CurrencyID int       `json:"currency_id"`
	Available  string    `json:"available"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
