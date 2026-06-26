package dto

import "time"

// REGISTER
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type RegisterResponse struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// LOGIN
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

// GET MY PROFILE
type GetMyProfileResponse struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// SUBSCRIBE
type SubscribeRequest struct {
	FolloweeID string `json:"followee_id"`
}

// UNSUBSCRIBE
type UnsubscribeRequest struct {
	FolloweeID string `json:"followee_id"`
}

// GET FOLLOWERS
type Follower struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

type GetFollowersRequest struct {
	UserID string `json:"user_id"`
}

type GetFollowersResponse struct {
	Followers []Follower `json:"followers"`
}
