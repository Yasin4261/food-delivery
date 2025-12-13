package domain

import (
	"errors"
	"math"
)

// Domain errors
var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrCannotCancelOrder       = errors.New("order cannot be cancelled")
	ErrInsufficientStock       = errors.New("insufficient stock")
	ErrStockNotManaged         = errors.New("stock is not managed for this item")
	ErrInvalidEmail            = errors.New("invalid email format")
	ErrInvalidPassword         = errors.New("invalid password")
	ErrUserNotFound            = errors.New("user not found")
	ErrEmailAlreadyExists      = errors.New("email already exists")
	ErrUsernameAlreadyExists   = errors.New("username already exists")
)

// CalculateDistance calculates distance between two coordinates in kilometers
// Uses Haversine formula
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers
	
	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)
	
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	return R * c
}

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
