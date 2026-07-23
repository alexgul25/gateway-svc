package dto

import "time"

// ADD PLACE
type AddPlaceRequest struct {
	Name string `json:"name"`
	Info string `json:"info"`
}

type AddPlaceResponse struct {
	PlaceID        string    `json:"place_id"`
	UserID         string    `json:"user_id"`
	PlaceName      string    `json:"place_name"`
	PlaceInfo      string    `json:"place_info"`
	PlaceCreatedAt time.Time `json:"place_created_at"`
}

// GET USER PLACES
type Place struct {
	Name      string    `json:"name"`
	Info      string    `json:"info"`
	CreatedAt time.Time `json:"created_at"`
}

type GetUserPlacesResponse struct {
	Places []Place `json:"places"`
}
