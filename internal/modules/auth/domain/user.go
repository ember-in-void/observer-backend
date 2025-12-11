package domain

import "time"

type UserID string

type User struct {
	ID        UserID
	Email     *string
	GoogleID  *string
	CreatedAt time.Time
	UpdatedAt time.Time
}
