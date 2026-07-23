package place

import "time"

type AddPlaceInfo struct {
	PlaceID        string
	UserID         string
	PlaceName      string
	PlaceInfo      string
	PlaceCreatedAt time.Time
}

type PlaceInfo struct {
	Name      string
	Info      string
	CreatedAt time.Time
}
