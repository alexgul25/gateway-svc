package user

import "time"

type RegisterInfo struct {
	UserID      string
	Email       string
	DisplayName string
	CreatedAt   time.Time
}

type GetMyProfileInfo struct {
	UserID      string
	Email       string
	DisplayName string
	CreatedAt   time.Time
}

type FollowerInfo struct {
	UserID      string
	Email       string
	DisplayName string
}
